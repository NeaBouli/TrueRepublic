package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

const ibcTwoChainSmokeEnv = "TRUEREPUBLIC_IBC_TWO_CHAIN_SMOKE"

const (
	ibcCompatibleRestartPhaseEnv    = "TRUEREPUBLIC_IBC_COMPATIBLE_RESTART_PHASE"
	ibcCompatibleRestartManifestEnv = "TRUEREPUBLIC_IBC_COMPATIBLE_RESTART_MANIFEST"
	ibcCompatibleCandidateVersion   = "gh181-compatible-candidate"
	ibcCompatibleFailMarker         = "gh181-intentional-fail-before-open"
)

type ibcCompatibleChainManifest struct {
	ChainID        string    `json:"chain_id"`
	Suffix         string    `json:"suffix"`
	SecretPrefix   string    `json:"secret_prefix"`
	DBDir          string    `json:"db_dir"`
	LastHeight     int64     `json:"last_height"`
	LastHeaderTime time.Time `json:"last_header_time"`
}

type ibcCompatibleEndpointManifest struct {
	ClientID     string `json:"client_id"`
	ConnectionID string `json:"connection_id"`
	ChannelID    string `json:"channel_id"`
}

type ibcCompatibleRestartManifest struct {
	Schema             int                           `json:"schema"`
	CurrentTime        time.Time                     `json:"current_time"`
	ChainA             ibcCompatibleChainManifest    `json:"chain_a"`
	ChainB             ibcCompatibleChainManifest    `json:"chain_b"`
	EndpointA          ibcCompatibleEndpointManifest `json:"endpoint_a"`
	EndpointB          ibcCompatibleEndpointManifest `json:"endpoint_b"`
	PendingPacket      channeltypes.Packet           `json:"pending_packet"`
	PendingAck         []byte                        `json:"pending_ack"`
	PendingAckCommit   []byte                        `json:"pending_ack_commitment"`
	FreshPacket        channeltypes.Packet           `json:"fresh_packet"`
	TimeoutPacket      channeltypes.Packet           `json:"timeout_packet"`
	SourceAddress      string                        `json:"source_address"`
	ReceiverAddress    string                        `json:"receiver_address"`
	PreSourceBalance   string                        `json:"pre_restart_source_balance"`
	PreEscrowBalance   string                        `json:"pre_restart_escrow_balance"`
	PreReceiverBalance string                        `json:"pre_restart_receiver_balance"`
	PreNextSequence    uint64                        `json:"pre_restart_next_sequence_send"`
	VoucherDenom       string                        `json:"voucher_denom"`
	SourceBalance      string                        `json:"source_balance"`
	EscrowBalance      string                        `json:"escrow_balance"`
	ReceiverBalance    string                        `json:"receiver_balance"`
	NextSequenceSend   uint64                        `json:"next_sequence_send"`
}

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
	chain        *ibctesting.TestChain
	suffix       string
	secretPrefix string
	dbDir        string
	homeDir      string
	db           dbm.DB
	staking      *trueRepublicIBCTestStakingKeeper
	validators   *cmttypes.ValidatorSet
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

func TestIBCTwoChainChannelCloseTimeoutRecoveryReplacement(t *testing.T) {
	if os.Getenv(ibcTwoChainSmokeEnv) != "1" {
		t.Skip("set " + ibcTwoChainSmokeEnv + "=1 to run the bounded two-chain IBC harness")
	}

	coord := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC),
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

	const closeAmount int64 = 900_000
	source := chainA.chain.SenderAccount.GetAddress()
	receiver := chainB.chain.SenderAccount.GetAddress()
	sourceBefore := chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount
	oldEscrow := transfertypes.GetEscrowAddress(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	packet := sendIBCTransfer(t, path, token.NewCoin(math.NewInt(closeAmount)), receiver.String(), 0)
	oldVoucherDenom := transfertypes.DenomTrace{Path: packet.DestinationPort + "/" + packet.DestinationChannel, BaseDenom: token.BaseDenom}.IBCDenom()
	require.Equal(t, sourceBefore.SubRaw(closeAmount), chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)
	require.Equal(t, math.NewInt(closeAmount), chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), oldEscrow, token.BaseDenom).Amount)
	require.True(t, chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), receiver, oldVoucherDenom).IsZero())
	require.NotEmpty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence))
	_, receiptFound := chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketReceipt(chainB.chain.GetContext(), packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	require.False(t, receiptFound)

	// ICS-20 deliberately rejects user-initiated channel close. Mirror ibc-go's
	// own TimeoutOnClose fixture by committing a counterparty CLOSED end, then
	// exercise the real proof-verified close-confirm and timeout messages.
	require.ErrorContains(t, path.EndpointB.ChanCloseInit(), "user cannot close channel")
	require.NoError(t, path.EndpointB.SetChannelState(channeltypes.CLOSED))
	restartTrueRepublicIBCChain(t, chainB, coord)
	path.EndpointB.Chain = chainB.chain
	path.EndpointA.Counterparty = path.EndpointB
	path.EndpointB.Counterparty = path.EndpointA
	requireIBCChannelState(t, path.EndpointB, channeltypes.CLOSED)
	require.NotEmpty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence))

	channelKey := ibchost.ChannelKey(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
	closedProof, proofHeight := path.EndpointB.QueryProof(channelKey)
	closeConfirm := channeltypes.NewMsgChannelCloseConfirm(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		closedProof,
		proofHeight,
		path.EndpointA.Chain.SenderAccount.GetAddress().String(),
	)
	_, err := path.EndpointA.Chain.SendMsgs(closeConfirm)
	require.NoError(t, err)
	requireIBCChannelState(t, path.EndpointA, channeltypes.CLOSED)

	require.NoError(t, path.EndpointA.TimeoutOnClose(packet))
	require.Equal(t, sourceBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)
	require.True(t, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), oldEscrow, token.BaseDenom).IsZero())
	require.Empty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence))
	require.NoError(t, path.EndpointA.TimeoutOnClose(packet), "duplicate timeout-on-close must be an IBC no-op")
	require.Equal(t, sourceBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)

	closedTransfer := transfertypes.NewMsgTransfer(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		token.NewCoin(math.NewInt(1)),
		source.String(),
		receiver.String(),
		clienttypes.NewHeight(clienttypes.ParseChainID(chainB.chain.ChainID), uint64(chainB.chain.App.LastBlockHeight()+100)),
		0,
		"gh178-closed-channel",
	)
	_, err = path.EndpointA.Chain.SendMsgs(closedTransfer)
	require.ErrorContains(t, err, "channel is not OPEN", "the old channel must reject specifically because it is closed")
	require.Equal(t, sourceBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)

	replacement := ibctesting.NewTransferPath(chainA.chain, chainB.chain)
	replacement.EndpointA.ClientID = path.EndpointA.ClientID
	replacement.EndpointA.ConnectionID = path.EndpointA.ConnectionID
	replacement.EndpointB.ClientID = path.EndpointB.ClientID
	replacement.EndpointB.ConnectionID = path.EndpointB.ConnectionID
	coord.CreateTransferChannels(replacement)
	requireIBCOpenUnorderedChannel(t, replacement)
	require.Equal(t, path.EndpointA.ConnectionID, replacement.EndpointA.ConnectionID)
	require.Equal(t, path.EndpointB.ConnectionID, replacement.EndpointB.ConnectionID)
	require.NotEqual(t, path.EndpointA.ChannelID, replacement.EndpointA.ChannelID)
	require.NotEqual(t, path.EndpointB.ChannelID, replacement.EndpointB.ChannelID)

	const replacementAmount int64 = 400_000
	replacementPacket, ack := sendAndReceiveIBCTransfer(t, replacement, token.NewCoin(math.NewInt(replacementAmount)), receiver.String(), 0)
	require.NoError(t, replacement.EndpointA.AcknowledgePacket(replacementPacket, ack))
	replacementEscrow := transfertypes.GetEscrowAddress(replacement.EndpointA.ChannelConfig.PortID, replacement.EndpointA.ChannelID)
	require.NotEqual(t, oldEscrow.String(), replacementEscrow.String())
	require.Equal(t, math.NewInt(replacementAmount), chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), replacementEscrow, token.BaseDenom).Amount)
	require.Equal(t, sourceBefore.SubRaw(replacementAmount), chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), source, token.BaseDenom).Amount)
	replacementVoucherDenom := transfertypes.DenomTrace{Path: replacementPacket.DestinationPort + "/" + replacementPacket.DestinationChannel, BaseDenom: token.BaseDenom}.IBCDenom()
	require.NotEqual(t, oldVoucherDenom, replacementVoucherDenom)
	require.Equal(t, math.NewInt(replacementAmount), chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), receiver, replacementVoucherDenom).Amount)
	require.True(t, chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), receiver, oldVoucherDenom).IsZero())
	require.Empty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), replacementPacket.SourcePort, replacementPacket.SourceChannel, replacementPacket.Sequence))
	requireIBCChannelState(t, path.EndpointA, channeltypes.CLOSED)
	requireIBCChannelState(t, path.EndpointB, channeltypes.CLOSED)

	chainA.chain.App.(*trueRepublicIBCTestingApp).crisisKeeper.AssertInvariants(chainA.chain.GetContext())
	chainB.chain.App.(*trueRepublicIBCTestingApp).crisisKeeper.AssertInvariants(chainB.chain.GetContext())
}

// TestIBCTwoChainCompatibleBinaryRestartRecovery proves only an in-place,
// state-compatible test-binary replacement. It is not x/upgrade evidence, a
// consensus migration, daemon/network relaying, or a compatibility guarantee
// for arbitrary source changes.
func TestIBCTwoChainCompatibleBinaryRestartRecovery(t *testing.T) {
	phase := os.Getenv(ibcCompatibleRestartPhaseEnv)
	if phase != "" {
		runIBCCompatibleRestartPhase(t, phase)
		return
	}
	if os.Getenv(ibcTwoChainSmokeEnv) != "1" {
		t.Skip("set " + ibcTwoChainSmokeEnv + "=1 to run the bounded two-chain IBC harness")
	}

	workDir := t.TempDir()
	manifestPath := filepath.Join(workDir, "manifest.json")
	baselineBinary, err := os.Executable()
	require.NoError(t, err)
	candidateBinary := filepath.Join(workDir, "candidate.test")
	build := exec.CommandContext(t.Context(), "go", "test", "-c", "-ldflags", "-X truerepublic.version="+ibcCompatibleCandidateVersion, "-o", candidateBinary, ".")
	buildOutput, err := build.CombinedOutput()
	require.NoError(t, err, "build compatible candidate test binary: %s", buildOutput)

	runIBCCompatibleRestartChild(t, baselineBinary, "baseline", manifestPath, true)
	manifest := readIBCCompatibleRestartManifest(t, manifestPath)
	before := hashIBCCompatibleStateDirs(t, manifest)
	failOutput := runIBCCompatibleRestartChild(t, candidateBinary, "fail-before-open", manifestPath, false)
	require.Contains(t, failOutput, ibcCompatibleFailMarker)
	require.Equal(t, before, hashIBCCompatibleStateDirs(t, manifest), "fail-before-open candidate must not mutate either application database")

	runIBCCompatibleRestartChild(t, candidateBinary, "candidate", manifestPath, true)
	runIBCCompatibleRestartChild(t, baselineBinary, "verify", manifestPath, true)
}

func runIBCCompatibleRestartPhase(t *testing.T, phase string) {
	t.Helper()
	manifestPath := os.Getenv(ibcCompatibleRestartManifestEnv)
	require.NotEmpty(t, manifestPath)
	switch phase {
	case "baseline":
		runIBCCompatibleBaselinePhase(t, manifestPath)
	case "fail-before-open":
		// The exact candidate artifact is already running, but deliberately exits
		// before reading the manifest or opening either LevelDB database.
		fmt.Fprintln(os.Stderr, ibcCompatibleFailMarker+" candidate="+version)
		os.Exit(42)
	case "candidate":
		require.Equal(t, ibcCompatibleCandidateVersion, version, "candidate binary must carry the independently linked version marker")
		runIBCCompatibleCandidatePhase(t, manifestPath)
	case "verify":
		runIBCCompatibleVerifyPhase(t, manifestPath)
	default:
		t.Fatalf("unsupported compatible restart phase %q", phase)
	}
}

func runIBCCompatibleBaselinePhase(t *testing.T, manifestPath string) {
	t.Helper()
	coord := &ibctesting.Coordinator{
		T: t, CurrentTime: time.Date(2026, 8, 12, 6, 0, 0, 0, time.UTC),
		Chains: make(map[string]*ibctesting.TestChain),
	}
	workDir := filepath.Dir(manifestPath)
	chainA := newTrueRepublicIBCChainAt(t, coord, ibctesting.GetChainID(1), "a", filepath.Join(workDir, "db-a"), "gh181")
	chainB := newTrueRepublicIBCChainAt(t, coord, ibctesting.GetChainID(2), "b", filepath.Join(workDir, "db-b"), "gh181")
	coord.Chains[chainA.chain.ChainID] = chainA.chain
	coord.Chains[chainB.chain.ChainID] = chainB.chain

	path := ibctesting.NewTransferPath(chainA.chain, chainB.chain)
	coord.SetupConnections(path)
	coord.CreateTransferChannels(path)
	requireIBCOpenUnorderedChannel(t, path)

	const pendingAmount int64 = 1_100_000
	receiver := chainB.chain.SenderAccount.GetAddress()
	packet, ack := sendAndReceiveIBCTransfer(t, path, token.NewCoin(math.NewInt(pendingAmount)), receiver.String(), 0)
	require.NotEmpty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence))
	_, receiptFound := chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketReceipt(chainB.chain.GetContext(), packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	require.True(t, receiptFound)
	storedAck, ackFound := chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketAcknowledgement(chainB.chain.GetContext(), packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	require.True(t, ackFound)
	require.NotEmpty(t, storedAck)

	escrow := transfertypes.GetEscrowAddress(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	voucherDenom := transfertypes.DenomTrace{Path: packet.DestinationPort + "/" + packet.DestinationChannel, BaseDenom: token.BaseDenom}.IBCDenom()
	nextSequence, nextSequenceFound := chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(chainA.chain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	require.True(t, nextSequenceFound)
	manifest := ibcCompatibleRestartManifest{
		Schema:      1,
		CurrentTime: coord.CurrentTime,
		ChainA:      ibcCompatibleChainState(chainA),
		ChainB:      ibcCompatibleChainState(chainB),
		EndpointA: ibcCompatibleEndpointManifest{
			ClientID: path.EndpointA.ClientID, ConnectionID: path.EndpointA.ConnectionID, ChannelID: path.EndpointA.ChannelID,
		},
		EndpointB: ibcCompatibleEndpointManifest{
			ClientID: path.EndpointB.ClientID, ConnectionID: path.EndpointB.ConnectionID, ChannelID: path.EndpointB.ChannelID,
		},
		PendingPacket:      packet,
		PendingAck:         ack,
		PendingAckCommit:   append([]byte(nil), storedAck...),
		SourceAddress:      chainA.chain.SenderAccount.GetAddress().String(),
		ReceiverAddress:    receiver.String(),
		VoucherDenom:       voucherDenom,
		PreSourceBalance:   chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount.String(),
		PreEscrowBalance:   chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount.String(),
		PreReceiverBalance: chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), receiver, voucherDenom).Amount.String(),
		PreNextSequence:    nextSequence,
	}
	writeIBCCompatibleRestartManifest(t, manifestPath, manifest)
	closeIBCCompatibleChain(t, chainA)
	closeIBCCompatibleChain(t, chainB)
}

func runIBCCompatibleCandidatePhase(t *testing.T, manifestPath string) {
	t.Helper()
	manifest := readIBCCompatibleRestartManifest(t, manifestPath)
	coord, chainA, chainB, path := reopenIBCCompatiblePair(t, manifest)

	requireIBCOpenUnorderedChannel(t, path)
	require.NotEmpty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), manifest.PendingPacket.SourcePort, manifest.PendingPacket.SourceChannel, manifest.PendingPacket.Sequence))
	_, receiptFound := chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketReceipt(chainB.chain.GetContext(), manifest.PendingPacket.DestinationPort, manifest.PendingPacket.DestinationChannel, manifest.PendingPacket.Sequence)
	require.True(t, receiptFound)
	storedAck, ackFound := chainB.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketAcknowledgement(chainB.chain.GetContext(), manifest.PendingPacket.DestinationPort, manifest.PendingPacket.DestinationChannel, manifest.PendingPacket.Sequence)
	require.True(t, ackFound)
	require.Equal(t, manifest.PendingAckCommit, storedAck)
	escrow := transfertypes.GetEscrowAddress(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	require.Equal(t, manifest.PreSourceBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount.String())
	require.Equal(t, manifest.PreEscrowBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount.String())
	require.Equal(t, manifest.PreReceiverBalance, chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), chainB.chain.SenderAccount.GetAddress(), manifest.VoucherDenom).Amount.String())
	preNextSequence, found := chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(chainA.chain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	require.True(t, found)
	require.Equal(t, manifest.PreNextSequence, preNextSequence)

	require.NoError(t, path.EndpointA.UpdateClient())
	require.NoError(t, path.EndpointA.AcknowledgePacket(manifest.PendingPacket, manifest.PendingAck))
	require.Empty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), manifest.PendingPacket.SourcePort, manifest.PendingPacket.SourceChannel, manifest.PendingPacket.Sequence))
	require.Equal(t, manifest.PreSourceBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount.String())
	require.Equal(t, manifest.PreEscrowBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount.String())
	require.Equal(t, manifest.PreReceiverBalance, chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), chainB.chain.SenderAccount.GetAddress(), manifest.VoucherDenom).Amount.String())
	require.NoError(t, path.EndpointA.AcknowledgePacket(manifest.PendingPacket, manifest.PendingAck), "duplicate post-restart acknowledgement must be an IBC no-op")
	require.Equal(t, manifest.PreSourceBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount.String())
	require.Equal(t, manifest.PreEscrowBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount.String())
	require.Equal(t, manifest.PreReceiverBalance, chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), chainB.chain.SenderAccount.GetAddress(), manifest.VoucherDenom).Amount.String())

	const freshAmount int64 = 500_000
	freshPacket, freshAck := sendAndReceiveIBCTransfer(t, path, token.NewCoin(math.NewInt(freshAmount)), manifest.ReceiverAddress, 0)
	require.NoError(t, path.EndpointA.AcknowledgePacket(freshPacket, freshAck))
	require.Empty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), freshPacket.SourcePort, freshPacket.SourceChannel, freshPacket.Sequence))
	freshSourceBalance := chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount
	freshEscrowBalance := chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount
	freshReceiverBalance := chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), chainB.chain.SenderAccount.GetAddress(), manifest.VoucherDenom).Amount
	require.NoError(t, path.EndpointA.AcknowledgePacket(freshPacket, freshAck), "duplicate fresh acknowledgement must be an IBC no-op")
	require.Equal(t, freshSourceBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount)
	require.Equal(t, freshEscrowBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount)
	require.Equal(t, freshReceiverBalance, chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), chainB.chain.SenderAccount.GetAddress(), manifest.VoucherDenom).Amount)

	const timeoutAmount int64 = 300_000
	timeoutSourceBefore := chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount
	timeoutEscrowBefore := chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount
	timeoutAt := uint64(coord.CurrentTime.Add(2 * time.Second).UnixNano())
	timeoutPacket := sendIBCTransfer(t, path, token.NewCoin(math.NewInt(timeoutAmount)), manifest.ReceiverAddress, timeoutAt)
	require.Equal(t, timeoutSourceBefore.SubRaw(timeoutAmount), chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount)
	nextBlockTime := coord.CurrentTime.Add(10 * time.Second)
	chainA.staking.recordTimeFromHeight(chainA.chain.App.LastBlockHeight()+1, nextBlockTime)
	chainB.staking.recordTimeFromHeight(chainB.chain.App.LastBlockHeight()+1, nextBlockTime)
	coord.IncrementTimeBy(10 * time.Second)
	coord.CommitBlock(chainB.chain)
	require.NoError(t, path.EndpointA.UpdateClient())
	require.NoError(t, path.EndpointA.TimeoutPacket(timeoutPacket))
	require.Equal(t, timeoutSourceBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount)
	require.Equal(t, timeoutEscrowBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount)
	require.NoError(t, path.EndpointA.TimeoutPacket(timeoutPacket), "duplicate post-restart timeout must be an IBC no-op")
	require.Equal(t, timeoutSourceBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount)
	require.Equal(t, timeoutEscrowBefore, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount)

	voucherDenom := manifest.VoucherDenom
	nextSequence, found := chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(chainA.chain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	require.True(t, found)
	manifest.CurrentTime = coord.CurrentTime
	manifest.ChainA = ibcCompatibleChainState(chainA)
	manifest.ChainB = ibcCompatibleChainState(chainB)
	manifest.FreshPacket = freshPacket
	manifest.TimeoutPacket = timeoutPacket
	manifest.VoucherDenom = voucherDenom
	manifest.SourceBalance = chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount.String()
	manifest.EscrowBalance = chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount.String()
	manifest.ReceiverBalance = chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), chainB.chain.SenderAccount.GetAddress(), voucherDenom).Amount.String()
	manifest.NextSequenceSend = nextSequence
	chainA.chain.App.(*trueRepublicIBCTestingApp).crisisKeeper.AssertInvariants(chainA.chain.GetContext())
	chainB.chain.App.(*trueRepublicIBCTestingApp).crisisKeeper.AssertInvariants(chainB.chain.GetContext())
	writeIBCCompatibleRestartManifest(t, manifestPath, manifest)
	closeIBCCompatibleChain(t, chainA)
	closeIBCCompatibleChain(t, chainB)
}

func runIBCCompatibleVerifyPhase(t *testing.T, manifestPath string) {
	t.Helper()
	manifest := readIBCCompatibleRestartManifest(t, manifestPath)
	_, chainA, chainB, path := reopenIBCCompatiblePair(t, manifest)
	requireIBCOpenUnorderedChannel(t, path)
	_, clientAFound := chainA.chain.App.GetIBCKeeper().ClientKeeper.GetClientState(chainA.chain.GetContext(), manifest.EndpointA.ClientID)
	_, clientBFound := chainB.chain.App.GetIBCKeeper().ClientKeeper.GetClientState(chainB.chain.GetContext(), manifest.EndpointB.ClientID)
	require.True(t, clientAFound)
	require.True(t, clientBFound)
	connectionA, connectionAFound := chainA.chain.App.GetIBCKeeper().ConnectionKeeper.GetConnection(chainA.chain.GetContext(), manifest.EndpointA.ConnectionID)
	connectionB, connectionBFound := chainB.chain.App.GetIBCKeeper().ConnectionKeeper.GetConnection(chainB.chain.GetContext(), manifest.EndpointB.ConnectionID)
	require.True(t, connectionAFound)
	require.True(t, connectionBFound)
	require.Equal(t, connectiontypes.OPEN, connectionA.State)
	require.Equal(t, connectiontypes.OPEN, connectionB.State)
	nextSequence, found := chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(chainA.chain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	require.True(t, found)
	require.Equal(t, manifest.NextSequenceSend, nextSequence)

	escrow := transfertypes.GetEscrowAddress(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	require.Equal(t, manifest.SourceBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), chainA.chain.SenderAccount.GetAddress(), token.BaseDenom).Amount.String())
	require.Equal(t, manifest.EscrowBalance, chainA.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainA.chain.GetContext(), escrow, token.BaseDenom).Amount.String())
	require.Equal(t, manifest.ReceiverBalance, chainB.chain.App.(*trueRepublicIBCTestingApp).bankKeeper.GetBalance(chainB.chain.GetContext(), chainB.chain.SenderAccount.GetAddress(), manifest.VoucherDenom).Amount.String())
	for _, packet := range []channeltypes.Packet{manifest.PendingPacket, manifest.FreshPacket, manifest.TimeoutPacket} {
		require.Empty(t, chainA.chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(chainA.chain.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence))
	}
	chainA.chain.App.(*trueRepublicIBCTestingApp).crisisKeeper.AssertInvariants(chainA.chain.GetContext())
	chainB.chain.App.(*trueRepublicIBCTestingApp).crisisKeeper.AssertInvariants(chainB.chain.GetContext())
	closeIBCCompatibleChain(t, chainA)
	closeIBCCompatibleChain(t, chainB)
}

func newTrueRepublicIBCChain(t *testing.T, coord *ibctesting.Coordinator, chainID, suffix string) *trueRepublicIBCChain {
	t.Helper()
	return newTrueRepublicIBCChainAt(t, coord, chainID, suffix, t.TempDir(), "gh175")
}

func newTrueRepublicIBCChainAt(t *testing.T, coord *ibctesting.Coordinator, chainID, suffix, dbDir, secretPrefix string) *trueRepublicIBCChain {
	t.Helper()
	require.NotEmpty(t, suffix)
	require.NotEmpty(t, dbDir)
	require.NotEmpty(t, secretPrefix)
	require.NoError(t, os.MkdirAll(dbDir, 0o700))
	homeDir := t.TempDir()
	database, err := dbm.NewDB("application", dbm.GoLevelDBBackend, dbDir)
	require.NoError(t, err)
	configureSDKConfig()

	validatorKey := cmted25519.GenPrivKeyFromSecret([]byte(secretPrefix + "-validator-" + suffix))
	privValidator := cmttypes.NewMockPVWithParams(validatorKey, false, false)
	validator := cmttypes.NewValidator(validatorKey.PubKey(), 1)
	validatorSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})
	signers := map[string]cmttypes.PrivValidator{validator.Address.String(): privValidator}

	senderKey := secp256k1.GenPrivKeyFromSecret([]byte(secretPrefix + "-sender-" + suffix))
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
	return &trueRepublicIBCChain{chain: chain, suffix: suffix, secretPrefix: secretPrefix, dbDir: dbDir, homeDir: homeDir, db: database, staking: staking, validators: validatorSet}
}

func reopenIBCCompatiblePair(t *testing.T, manifest ibcCompatibleRestartManifest) (*ibctesting.Coordinator, *trueRepublicIBCChain, *trueRepublicIBCChain, *ibctesting.Path) {
	t.Helper()
	require.Equal(t, 1, manifest.Schema)
	coord := &ibctesting.Coordinator{T: t, CurrentTime: manifest.CurrentTime, Chains: make(map[string]*ibctesting.TestChain)}
	chainA := reopenIBCCompatibleChain(t, coord, manifest.ChainA)
	chainB := reopenIBCCompatibleChain(t, coord, manifest.ChainB)
	coord.Chains[chainA.chain.ChainID] = chainA.chain
	coord.Chains[chainB.chain.ChainID] = chainB.chain
	path := ibctesting.NewTransferPath(chainA.chain, chainB.chain)
	path.EndpointA.ClientID = manifest.EndpointA.ClientID
	path.EndpointA.ConnectionID = manifest.EndpointA.ConnectionID
	path.EndpointA.ChannelID = manifest.EndpointA.ChannelID
	path.EndpointB.ClientID = manifest.EndpointB.ClientID
	path.EndpointB.ConnectionID = manifest.EndpointB.ConnectionID
	path.EndpointB.ChannelID = manifest.EndpointB.ChannelID
	return coord, chainA, chainB, path
}

func reopenIBCCompatibleChain(t *testing.T, coord *ibctesting.Coordinator, manifest ibcCompatibleChainManifest) *trueRepublicIBCChain {
	t.Helper()
	require.NotEmpty(t, manifest.DBDir)
	database, err := dbm.NewDB("application", dbm.GoLevelDBBackend, manifest.DBDir)
	require.NoError(t, err)
	configureSDKConfig()

	require.NotEmpty(t, manifest.SecretPrefix)
	validatorKey := cmted25519.GenPrivKeyFromSecret([]byte(manifest.SecretPrefix + "-validator-" + manifest.Suffix))
	privValidator := cmttypes.NewMockPVWithParams(validatorKey, false, false)
	validator := cmttypes.NewValidator(validatorKey.PubKey(), 1)
	validatorSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})
	signers := map[string]cmttypes.PrivValidator{validator.Address.String(): privValidator}
	senderKey := secp256k1.GenPrivKeyFromSecret([]byte(manifest.SecretPrefix + "-sender-" + manifest.Suffix))

	staking := newIBCTestStakingKeeper(t, validatorSet, manifest.LastHeaderTime)
	homeDir := t.TempDir()
	app := NewTrueRepublicApp(log.NewNopLogger(), database, homeDir)
	baseapp.SetChainID(manifest.ChainID)(app.BaseApp)
	require.Equal(t, manifest.LastHeight, app.LastBlockHeight(), "candidate must reopen in-place state without InitChain or height reset")
	wrapper := &trueRepublicIBCTestingApp{TrueRepublicApp: app, staking: staking}
	app.ibcKeeper.SetConsensusHost(ibctm.NewConsensusHost(staking))

	chain := &ibctesting.TestChain{
		TB: t, Coordinator: coord, ChainID: manifest.ChainID, App: wrapper,
		QueryServer: app.ibcKeeper, TxConfig: app.txConfig, Codec: app.appCodec,
		Vals: validatorSet, NextVals: validatorSet, Signers: signers,
		SenderPrivKey: senderKey,
	}
	configureIBCHistoricalHeaderProvider(t, staking, app, manifest.ChainID, validatorSet)
	lastHeader, found := staking.header(manifest.LastHeight)
	require.True(t, found)
	chain.CurrentHeader = lastHeader
	chain.LastHeader = chain.CurrentTMClientHeader()
	chain.CurrentHeader = cmtproto.Header{
		ChainID: manifest.ChainID, Height: app.LastBlockHeight() + 1,
		Time: coord.CurrentTime, AppHash: app.LastCommitID().Hash,
		ValidatorsHash: validatorSet.Hash(), NextValidatorsHash: validatorSet.Hash(),
	}
	account := app.accountKeeper.GetAccount(chain.GetContext(), sdk.AccAddress(senderKey.PubKey().Address()))
	require.NotNil(t, account)
	chain.SenderAccount = account
	chain.SenderAccounts = []ibctesting.SenderAccount{{SenderPrivKey: senderKey, SenderAccount: account}}
	return &trueRepublicIBCChain{chain: chain, suffix: manifest.Suffix, secretPrefix: manifest.SecretPrefix, dbDir: manifest.DBDir, homeDir: homeDir, db: database, staking: staking, validators: validatorSet}
}

func ibcCompatibleChainState(chain *trueRepublicIBCChain) ibcCompatibleChainManifest {
	lastHeaderTime := chain.chain.CurrentHeader.Time
	if chain.chain.LastHeader != nil && chain.chain.LastHeader.SignedHeader != nil && chain.chain.LastHeader.SignedHeader.Header != nil {
		lastHeaderTime = chain.chain.LastHeader.SignedHeader.Header.Time
	}
	return ibcCompatibleChainManifest{
		ChainID: chain.chain.ChainID, Suffix: chain.suffix, SecretPrefix: chain.secretPrefix, DBDir: chain.dbDir,
		LastHeight: chain.chain.App.LastBlockHeight(), LastHeaderTime: lastHeaderTime,
	}
}

func closeIBCCompatibleChain(t *testing.T, chain *trueRepublicIBCChain) {
	t.Helper()
	if chain.db == nil {
		return
	}
	require.NoError(t, chain.chain.App.(*trueRepublicIBCTestingApp).Close())
	chain.db = nil
}

func runIBCCompatibleRestartChild(t *testing.T, binary, phase, manifestPath string, expectSuccess bool) string {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), binary,
		"-test.run", "^TestIBCTwoChainCompatibleBinaryRestartRecovery$",
		"-test.count=1", "-test.timeout=2m", "-test.v",
	)
	cmd.Env = append(os.Environ(),
		ibcTwoChainSmokeEnv+"=1",
		ibcCompatibleRestartPhaseEnv+"="+phase,
		ibcCompatibleRestartManifestEnv+"="+manifestPath,
	)
	output, err := cmd.CombinedOutput()
	if expectSuccess {
		require.NoError(t, err, "%s child failed: %s", phase, output)
	} else {
		require.Error(t, err, "%s child must fail deterministically", phase)
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok)
		require.Equal(t, 42, exitErr.ExitCode())
	}
	return string(output)
}

func readIBCCompatibleRestartManifest(t *testing.T, path string) ibcCompatibleRestartManifest {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	var manifest ibcCompatibleRestartManifest
	require.NoError(t, json.Unmarshal(content, &manifest))
	require.Equal(t, 1, manifest.Schema)
	return manifest
}

func writeIBCCompatibleRestartManifest(t *testing.T, path string, manifest ibcCompatibleRestartManifest) {
	t.Helper()
	content, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	temporary := path + ".tmp"
	require.NoError(t, os.WriteFile(temporary, content, 0o600))
	require.NoError(t, os.Rename(temporary, path))
}

func hashIBCCompatibleStateDirs(t *testing.T, manifest ibcCompatibleRestartManifest) map[string]string {
	t.Helper()
	return map[string]string{
		"chain_a": hashIBCCompatibleDirectory(t, manifest.ChainA.DBDir),
		"chain_b": hashIBCCompatibleDirectory(t, manifest.ChainB.DBDir),
	}
}

func hashIBCCompatibleDirectory(t *testing.T, root string) string {
	t.Helper()
	var paths []string
	require.NoError(t, filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			paths = append(paths, path)
		}
		return nil
	}))
	sort.Strings(paths)
	digest := sha256.New()
	for _, path := range paths {
		relative, err := filepath.Rel(root, path)
		require.NoError(t, err)
		_, err = io.WriteString(digest, relative+"\x00")
		require.NoError(t, err)
		file, err := os.Open(path)
		require.NoError(t, err)
		_, copyErr := io.Copy(digest, file)
		closeErr := file.Close()
		require.NoError(t, copyErr)
		require.NoError(t, closeErr)
	}
	return fmt.Sprintf("%x", digest.Sum(nil))
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

func requireIBCChannelState(t *testing.T, endpoint *ibctesting.Endpoint, expected channeltypes.State) {
	t.Helper()
	channel, found := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetChannel(endpoint.Chain.GetContext(), endpoint.ChannelConfig.PortID, endpoint.ChannelID)
	require.True(t, found)
	require.Equal(t, expected, channel.State)
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
