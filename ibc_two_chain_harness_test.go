package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cryptoproto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

const ibcTwoChainSmokeEnv = "TRUEREPUBLIC_IBC_TWO_CHAIN_SMOKE"

type trueRepublicIBCTestStakingKeeper struct {
	validators  stakingtypes.Validators
	header      func(int64) (cmtproto.Header, bool)
	initialTime time.Time
	timeChanges map[int64]time.Time
}

type ibcCommitInfoReader interface {
	GetCommitInfo(int64) (*storetypes.CommitInfo, error)
}

func (k *trueRepublicIBCTestStakingKeeper) GetHistoricalInfo(_ context.Context, height int64) (stakingtypes.HistoricalInfo, error) {
	if height < 1 {
		return stakingtypes.HistoricalInfo{}, fmt.Errorf("invalid historical height %d", height)
	}
	header, found := k.header(height)
	if !found {
		return stakingtypes.HistoricalInfo{}, fmt.Errorf("historical header %d not captured", height)
	}
	return stakingtypes.HistoricalInfo{Header: header, Valset: k.validators.Validators}, nil
}

func (*trueRepublicIBCTestStakingKeeper) UnbondingTime(context.Context) (time.Duration, error) {
	return ibctesting.UnbondingPeriod, nil
}

func (k *trueRepublicIBCTestStakingKeeper) recordTimeFromHeight(height int64, at time.Time) {
	k.timeChanges[height] = at
}

func (k *trueRepublicIBCTestStakingKeeper) timeAtHeight(height int64) time.Time {
	selectedHeight := int64(0)
	selectedTime := k.initialTime
	for fromHeight, at := range k.timeChanges {
		if fromHeight <= height && fromHeight >= selectedHeight {
			selectedHeight = fromHeight
			selectedTime = at
		}
	}
	return selectedTime
}

type trueRepublicIBCTestingApp struct {
	*TrueRepublicApp
	staking *trueRepublicIBCTestStakingKeeper
}

func (a *trueRepublicIBCTestingApp) GetBaseApp() *baseapp.BaseApp { return a.BaseApp }
func (a *trueRepublicIBCTestingApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return a.staking
}
func (a *trueRepublicIBCTestingApp) GetIBCKeeper() *ibckeeper.Keeper { return a.ibcKeeper }
func (*trueRepublicIBCTestingApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	// The real transfer/channel keepers already own their sealed scoped
	// capabilities. The proof-driven harness never calls the testing package's
	// direct capability helpers, so exposing a second scope would be unsafe.
	return capabilitykeeper.ScopedKeeper{}
}
func (a *trueRepublicIBCTestingApp) GetTxConfig() client.TxConfig { return a.txConfig }
func (a *trueRepublicIBCTestingApp) AppCodec() codec.Codec        { return a.appCodec }

var _ ibctesting.TestingApp = (*trueRepublicIBCTestingApp)(nil)

type trueRepublicIBCChain struct {
	chain      *ibctesting.TestChain
	dbDir      string
	homeDir    string
	db         dbm.DB
	staking    *trueRepublicIBCTestStakingKeeper
	validators *cmttypes.ValidatorSet
}

func TestIBCTwoChainTransferAcknowledgementTimeoutReplayRecovery(t *testing.T) {
	if os.Getenv(ibcTwoChainSmokeEnv) != "1" {
		t.Skip("set " + ibcTwoChainSmokeEnv + "=1 to run the bounded two-chain IBC harness")
	}

	coord := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
		Chains:      make(map[string]*ibctesting.TestChain),
	}
	chainA := newTrueRepublicIBCChain(t, coord, ibctesting.GetChainID(1), "a")
	chainB := newTrueRepublicIBCChain(t, coord, ibctesting.GetChainID(2), "b")
	coord.Chains[chainA.chain.ChainID] = chainA.chain
	coord.Chains[chainB.chain.ChainID] = chainB.chain
	t.Cleanup(func() {
		if chainA.db != nil {
			_ = chainA.chain.App.(*trueRepublicIBCTestingApp).Close()
		}
		if chainB.db != nil {
			_ = chainB.chain.App.(*trueRepublicIBCTestingApp).Close()
		}
	})

	path := ibctesting.NewTransferPath(chainA.chain, chainB.chain)
	coord.SetupConnections(path)
	coord.CreateTransferChannels(path)
	requireIBCOpenUnorderedChannel(t, path)

	const transferAmount int64 = 1_250_000
	source := chainA.chain.SenderAccount.GetAddress()
	receiver := chainB.chain.SenderAccount.GetAddress()
	sourceBefore := chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount
	escrow := transfertypes.GetEscrowAddress(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)

	packet, ack := sendAndReceiveIBCTransfer(t, path, token.NewCoin(math.NewInt(transferAmount)), receiver.String(), 0)
	require.Equal(t, sourceBefore.SubRaw(transferAmount), chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)
	require.Equal(t, math.NewInt(transferAmount), chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount)
	voucherDenom := transfertypes.DenomTrace{Path: packet.DestinationPort + "/" + packet.DestinationChannel, BaseDenom: token.BaseDenom}.IBCDenom()
	require.Equal(t, math.NewInt(transferAmount), chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), receiver, voucherDenom).Amount)
	require.NotEmpty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence))
	_, receiptFound := chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketReceipt(chainB.chain.GetContext(), packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	require.True(t, receiptFound)
	storedAck, ackFound := chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketAcknowledgement(chainB.chain.GetContext(), packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	require.True(t, ackFound)
	require.NotEmpty(t, storedAck)
	ackCommitment := append([]byte(nil), storedAck...)

	voucherBeforeReplay := chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), receiver, voucherDenom).Amount
	// Refresh the destination's client first so the replay reaches ibc-go's
	// receipt check instead of failing early on a stale proof height.
	require.NoError(t, path.EndpointB.UpdateClient())
	replayResult, err := path.EndpointB.RecvPacketWithResult(packet)
	require.NoError(t, err, "duplicate receive must be an IBC no-op")
	require.Zero(t, replayResult.Code)
	require.Equal(t, voucherBeforeReplay, chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), receiver, voucherDenom).Amount)

	// Reopen the destination database while the acknowledgement is pending.
	restartTrueRepublicIBCChain(t, chainB, coord)
	path.EndpointB.Chain = chainB.chain
	path.EndpointA.Counterparty = path.EndpointB
	path.EndpointB.Counterparty = path.EndpointA
	require.NoError(t, path.EndpointA.UpdateClient(), "recovered destination must commit and verify a fresh header")
	_, receiptFound = chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketReceipt(chainB.chain.GetContext(), packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	require.True(t, receiptFound)
	storedAck, ackFound = chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketAcknowledgement(chainB.chain.GetContext(), packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	require.True(t, ackFound)
	require.Equal(t, ackCommitment, storedAck)

	require.NoError(t, path.EndpointA.AcknowledgePacket(packet, ack))
	require.Empty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence))
	escrowAfterAck := chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount
	require.Equal(t, math.NewInt(transferAmount), escrowAfterAck)
	require.NoError(t, path.EndpointA.AcknowledgePacket(packet, ack), "duplicate acknowledgement is an IBC no-op")
	require.Equal(t, escrowAfterAck, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount)

	const timeoutAmount int64 = 750_000
	timeoutSourceBefore := chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount
	timeoutAt := uint64(coord.CurrentTime.Add(2 * time.Second).UnixNano())
	timeoutPacket := sendIBCTransfer(t, path, token.NewCoin(math.NewInt(timeoutAmount)), receiver.String(), timeoutAt)
	require.Equal(t, timeoutSourceBefore.SubRaw(timeoutAmount), chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)

	nextBlockTime := coord.CurrentTime.Add(10 * time.Second)
	chainA.staking.recordTimeFromHeight(chainA.chain.App.LastBlockHeight()+1, nextBlockTime)
	chainB.staking.recordTimeFromHeight(chainB.chain.App.LastBlockHeight()+1, nextBlockTime)
	coord.IncrementTimeBy(10 * time.Second)
	coord.CommitBlock(chainB.chain)
	require.NoError(t, path.EndpointA.UpdateClient())
	require.NoError(t, path.EndpointA.TimeoutPacket(timeoutPacket))
	require.Equal(t, timeoutSourceBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)
	require.Empty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), timeoutPacket.SourcePort, timeoutPacket.SourceChannel, timeoutPacket.Sequence))
	require.NoError(t, path.EndpointA.TimeoutPacket(timeoutPacket), "duplicate timeout is an IBC no-op")
	require.Equal(t, timeoutSourceBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)

	chainA.chain.App.(*trueRepublicIBCTestingApp).crisisKeeper.AssertInvariants(chainA.chain.GetContext())
	chainB.chain.App.(*trueRepublicIBCTestingApp).crisisKeeper.AssertInvariants(chainB.chain.GetContext())
}

func newTrueRepublicIBCChain(t *testing.T, coord *ibctesting.Coordinator, chainID, suffix string) *trueRepublicIBCChain {
	t.Helper()
	require.NotEmpty(t, suffix)
	dbDir := t.TempDir()
	homeDir := t.TempDir()
	database, err := dbm.NewDB("application", dbm.GoLevelDBBackend, dbDir)
	require.NoError(t, err)
	configureSDKConfig()

	validatorKey := cmted25519.GenPrivKeyFromSecret([]byte("gh175-validator-" + suffix))
	privValidator := cmttypes.NewMockPVWithParams(validatorKey, false, false)
	validator := cmttypes.NewValidator(validatorKey.PubKey(), 1)
	validatorSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})
	signers := map[string]cmttypes.PrivValidator{validator.Address.String(): privValidator}

	senderKey := secp256k1.GenPrivKeyFromSecret([]byte("gh175-sender-" + suffix))
	senderAddress := sdk.AccAddress(senderKey.PubKey().Address())
	senderAddressString := senderAddress.String()
	require.NotEmpty(t, senderAddressString)
	senderAccount := authtypes.NewBaseAccount(senderAddress, senderKey.PubKey(), 0, 0)
	staking := newIBCTestStakingKeeper(t, validatorSet, coord.CurrentTime)
	app := NewTrueRepublicApp(log.NewNopLogger(), database, homeDir)
	baseapp.SetChainID(chainID)(app.BaseApp)
	wrapper := &trueRepublicIBCTestingApp{TrueRepublicApp: app, staking: staking}
	app.ibcKeeper.SetConsensusHost(ibctm.NewConsensusHost(staking))

	state := defaultGenesisForApp(app)
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), []authtypes.GenesisAccount{senderAccount})
	state[authtypes.ModuleName] = app.appCodec.MustMarshalJSON(authGenesis)
	setBankGenesis(t, app, state, []banktypes.Balance{{Address: senderAddressString, Coins: sdk.NewCoins(token.NewCoin(math.NewInt(2_000_000_000)))}})
	democracyGenesis := truedemocracy.DefaultGenesisState()
	democracyGenesis.BootstrapOperatorAddresses = []string{sdk.AccAddress(bytes.Repeat([]byte{suffix[0]}, 20)).String()}
	democracyJSON, err := json.Marshal(democracyGenesis)
	require.NoError(t, err)
	state[truedemocracy.ModuleName] = democracyJSON
	validatorUpdates := []abci.ValidatorUpdate{{PubKey: cryptoproto.PublicKey{Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: validatorKey.PubKey().Bytes()}}, Power: 1}}
	require.NoError(t, ensureConsensusGenesis(app.appCodec, state, validatorUpdates))
	for _, balance := range banktypes.GetGenesisStateFromAppState(app.appCodec, state).Balances {
		require.NotEmpty(t, balance.Address, "bank genesis contains an empty address")
	}
	stateBytes, err := json.Marshal(state)
	require.NoError(t, err)
	consensusParams := cmttypes.DefaultConsensusParams().ToProto()
	_, err = app.InitChain(&abci.RequestInitChain{
		ChainId: chainID, Time: coord.CurrentTime, AppStateBytes: stateBytes,
		Validators:      validatorUpdates,
		ConsensusParams: &consensusParams,
	})
	require.NoError(t, err)

	chain := &ibctesting.TestChain{
		TB: t, Coordinator: coord, ChainID: chainID, App: wrapper,
		CurrentHeader: cmtproto.Header{ChainID: chainID, Height: 1, Time: coord.CurrentTime},
		QueryServer:   app.ibcKeeper, TxConfig: app.txConfig, Codec: app.appCodec,
		Vals: validatorSet, NextVals: validatorSet, Signers: signers,
		SenderPrivKey: senderKey, SenderAccount: senderAccount,
		SenderAccounts: []ibctesting.SenderAccount{{SenderPrivKey: senderKey, SenderAccount: senderAccount}},
	}
	configureIBCHistoricalHeaderProvider(t, staking, app, chainID, validatorSet)
	chain.NextBlock()
	return &trueRepublicIBCChain{chain: chain, dbDir: dbDir, homeDir: homeDir, db: database, staking: staking, validators: validatorSet}
}

func newIBCTestStakingKeeper(t *testing.T, validatorSet *cmttypes.ValidatorSet, initialTime time.Time) *trueRepublicIBCTestStakingKeeper {
	t.Helper()
	validators := make([]stakingtypes.Validator, 0, len(validatorSet.Validators))
	for _, validator := range validatorSet.Validators {
		pubKey, err := cryptocodec.FromCmtPubKeyInterface(validator.PubKey)
		require.NoError(t, err)
		validatorState, err := stakingtypes.NewValidator(sdk.ValAddress(validator.Address).String(), pubKey, stakingtypes.Description{})
		require.NoError(t, err)
		validatorState.Status = stakingtypes.Bonded
		validatorState.Tokens = sdk.DefaultPowerReduction
		validatorState.DelegatorShares = math.LegacyOneDec()
		validators = append(validators, validatorState)
	}
	return &trueRepublicIBCTestStakingKeeper{
		validators:  stakingtypes.Validators{Validators: validators},
		initialTime: initialTime,
		timeChanges: make(map[int64]time.Time),
	}
}

func restartTrueRepublicIBCChain(t *testing.T, state *trueRepublicIBCChain, coord *ibctesting.Coordinator) {
	t.Helper()
	old := state.chain.App.(*trueRepublicIBCTestingApp)
	require.NoError(t, old.Close())
	state.db = nil
	database, err := dbm.NewDB("application", dbm.GoLevelDBBackend, state.dbDir)
	require.NoError(t, err)
	// The IBC application state is in the reopened database. Use a fresh empty
	// wasm cache directory so the in-process predecessor's VM file lock cannot
	// masquerade as an IBC recovery failure.
	state.homeDir = t.TempDir()
	app := NewTrueRepublicApp(log.NewNopLogger(), database, state.homeDir)
	baseapp.SetChainID(state.chain.ChainID)(app.BaseApp)
	wrapper := &trueRepublicIBCTestingApp{TrueRepublicApp: app, staking: state.staking}
	app.ibcKeeper.SetConsensusHost(ibctm.NewConsensusHost(state.staking))
	state.chain.App = wrapper
	state.chain.QueryServer = app.ibcKeeper
	state.chain.TxConfig = app.txConfig
	state.chain.Codec = app.appCodec
	state.chain.CurrentHeader.AppHash = app.LastCommitID().Hash
	state.chain.CurrentHeader.Height = app.LastBlockHeight() + 1
	state.chain.CurrentHeader.Time = coord.CurrentTime
	configureIBCHistoricalHeaderProvider(t, state.staking, app, state.chain.ChainID, state.validators)
	state.db = database
}

func configureIBCHistoricalHeaderProvider(
	t *testing.T,
	staking *trueRepublicIBCTestStakingKeeper,
	app *TrueRepublicApp,
	chainID string,
	validatorSet *cmttypes.ValidatorSet,
) {
	t.Helper()
	commitInfoReader, ok := app.CommitMultiStore().(ibcCommitInfoReader)
	require.True(t, ok, "application commit store must expose historical commit information")
	historicalHeaders := make(map[int64]cmtproto.Header)
	staking.header = func(height int64) (cmtproto.Header, bool) {
		if header, found := historicalHeaders[height]; found {
			return header, true
		}
		var appHash []byte
		if height > 1 {
			commitInfo, err := commitInfoReader.GetCommitInfo(height - 1)
			if err != nil {
				return cmtproto.Header{}, false
			}
			appHash = commitInfo.Hash()
		}
		header := cmtproto.Header{
			ChainID:            chainID,
			Height:             height,
			Time:               staking.timeAtHeight(height),
			AppHash:            appHash,
			ValidatorsHash:     validatorSet.Hash(),
			NextValidatorsHash: validatorSet.Hash(),
		}
		historicalHeaders[height] = header
		return header, true
	}
}

func requireIBCOpenUnorderedChannel(t *testing.T, path *ibctesting.Path) {
	t.Helper()
	for _, endpoint := range []*ibctesting.Endpoint{path.EndpointA, path.EndpointB} {
		channel, found := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetChannel(endpoint.Chain.GetContext(), endpoint.ChannelConfig.PortID, endpoint.ChannelID)
		require.True(t, found)
		require.Equal(t, channeltypes.OPEN, channel.State)
		require.Equal(t, channeltypes.UNORDERED, channel.Ordering)
		require.Equal(t, transfertypes.Version, channel.Version)
	}
}

func sendIBCTransfer(t *testing.T, path *ibctesting.Path, coin sdk.Coin, receiver string, timeoutTimestamp uint64) channeltypes.Packet {
	t.Helper()
	msg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coin,
		path.EndpointA.Chain.SenderAccount.GetAddress().String(), receiver,
		clienttypes.NewHeight(clienttypes.ParseChainID(path.EndpointB.Chain.ChainID), uint64(path.EndpointB.Chain.App.LastBlockHeight()+100)), timeoutTimestamp, "gh175")
	result, err := path.EndpointA.Chain.SendMsgs(msg)
	require.NoError(t, err)
	packet, err := ibctesting.ParsePacketFromEvents(result.Events)
	require.NoError(t, err)
	return packet
}

func sendAndReceiveIBCTransfer(t *testing.T, path *ibctesting.Path, coin sdk.Coin, receiver string, timeoutTimestamp uint64) (channeltypes.Packet, []byte) {
	t.Helper()
	packet := sendIBCTransfer(t, path, coin, receiver, timeoutTimestamp)
	require.NoError(t, path.EndpointB.UpdateClient())
	receiveResult, err := path.EndpointB.RecvPacketWithResult(packet)
	require.NoError(t, err)
	ack, err := ibctesting.ParseAckFromEvents(receiveResult.Events)
	require.NoError(t, err)
	return packet, ack
}
