package main

import (
	"encoding/hex"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"

	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

// TestClientCustomTxVectorsMatchGoWireEncoding is the chain-side half of the
// deterministic vectors in client-web/src/services/txRegistry.test.ts. It
// prevents either hand-written browser codecs or Go gogoproto identities from
// drifting without a cross-language test failure.
func TestClientCustomTxVectorsMatchGoWireEncoding(t *testing.T) {
	coin := func(amount int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewCoin("upnyx", math.NewInt(amount)))
	}
	tests := []struct {
		typeURL string
		msg     gogoproto.Message
		hex     string
	}{
		{
			typeURL: "/truedemocracy.MsgCreateDomain",
			msg: &truedemocracy.MsgCreateDomain{
				Name: "dom", Admin: sdk.AccAddress{1, 2, 3, 4}, InitialCoins: coin(42),
			},
			hex: "0a03646f6d1204010203041a0b0a0575706e797812023432",
		},
		{
			typeURL: "/truedemocracy.MsgSubmitProposal",
			msg: &truedemocracy.MsgSubmitProposal{
				Sender: sdk.AccAddress{9, 8}, DomainName: "d", IssueName: "i",
				SuggestionName: "s", Creator: "c", Fee: coin(7), ExternalLink: "L",
			},
			hex: "0a0209081201641a01692201732a0163320a0a0575706e79781201373a014c",
		},
		{
			typeURL: "/truedemocracy.MsgPlaceStoneOnSuggestion",
			msg: &truedemocracy.MsgPlaceStoneOnSuggestion{
				Sender: sdk.AccAddress{1}, DomainName: "d", IssueName: "i",
				SuggestionName: "s", MemberAddr: "m",
			},
			hex: "0a01011201641a01692201732a016d",
		},
		{
			typeURL: "/truedemocracy.MsgPlaceStoneOnIssue",
			msg: &truedemocracy.MsgPlaceStoneOnIssue{
				Sender: sdk.AccAddress{1}, DomainName: "d", IssueName: "i", MemberAddr: "m",
			},
			hex: "0a01011201641a016922016d",
		},
		{
			typeURL: "/truedemocracy.MsgApproveOnboarding",
			msg: &truedemocracy.MsgApproveOnboarding{
				Sender: sdk.AccAddress{1}, DomainName: "d", RequesterAddr: "r",
			},
			hex: "0a01011201641a0172",
		},
		{
			typeURL: "/truedemocracy.MsgAddMember",
			msg: &truedemocracy.MsgAddMember{
				Sender: sdk.AccAddress{1}, DomainName: "d", NewMember: "n",
			},
			hex: "0a01011201641a016e",
		},
		{
			typeURL: "/truedemocracy.MsgOnboardToDomain",
			msg: &truedemocracy.MsgOnboardToDomain{
				Sender: sdk.AccAddress{1}, DomainName: "d", DomainPubKeyHex: "dp",
				GlobalPubKeyHex: "gp", SignatureHex: "sg",
			},
			hex: "0a01011201641a026470220267702a027367",
		},
		{
			typeURL: "/truedemocracy.MsgRegisterIdentity",
			msg: &truedemocracy.MsgRegisterIdentity{
				Sender: sdk.AccAddress{1}, DomainName: "d", Commitment: "ci",
			},
			hex: "0a01011201641a026369",
		},
		{
			typeURL: "/dex.MsgAddLiquidity",
			msg: &dex.MsgAddLiquidity{
				Sender: sdk.AccAddress{1}, AssetDenom: "atom", PnyxAmt: 300, AssetAmt: 1,
			},
			hex: "0a0101120461746f6d18ac022001",
		},
		{
			typeURL: "/dex.MsgRemoveLiquidity",
			msg: &dex.MsgRemoveLiquidity{
				Sender: sdk.AccAddress{1}, AssetDenom: "atom", Shares: 5,
			},
			hex: "0a0101120461746f6d1805",
		},
		{
			typeURL: "/dex.MsgSwapExact",
			msg: &dex.MsgSwapExact{
				Sender: sdk.AccAddress{1}, InputDenom: "upnyx", InputAmt: 2,
				OutputDenom: "atom", MinOutput: 1,
			},
			hex: "0a0101120575706e79781802220461746f6d2801",
		},
	}

	for _, tc := range tests {
		t.Run(tc.typeURL, func(t *testing.T) {
			if got := "/" + gogoproto.MessageName(tc.msg); got != tc.typeURL {
				t.Fatalf("registered type URL = %q, want %q", got, tc.typeURL)
			}
			encoded, err := gogoproto.Marshal(tc.msg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if got := hex.EncodeToString(encoded); got != tc.hex {
				t.Fatalf("wire bytes = %s, want %s", got, tc.hex)
			}
		})
	}
}
