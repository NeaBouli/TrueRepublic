package dex

import (
	"fmt"

	txsigning "cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type legacyMsgSigner interface {
	GetSigners() []sdk.AccAddress
}

// RegisterCustomGetSigners wires the hand-written DEX messages into the
// Cosmos SDK v0.50 signing context. DEX messages expose sdk.Msg.GetSigners,
// but do not have generated cosmos.msg.v1.signer annotations.
func RegisterCustomGetSigners(options *txsigning.Options) {
	signerFields := msgSignerFields()
	for _, msgType := range msgTypesForDescriptor() {
		typeName := protoreflect.FullName("dex." + msgType.Elem().Name())
		signerField, ok := signerFields[msgType]
		if !ok {
			panic("missing signer contract for " + typeName)
		}
		options.DefineCustomGetSigners(typeName, func(msg proto.Message) ([][]byte, error) {
			if legacyMsg, ok := msg.(legacyMsgSigner); ok {
				return accAddressSignersToBytes(legacyMsg.GetSigners()), nil
			}

			// The SDK converts hand-written gogo messages to dynamicpb messages
			// while decoding TxRaw. Resolve the canonical sender field by
			// reflection so signer extraction works on both representations.
			reflectedMsg := msg.ProtoReflect()
			field := reflectedMsg.Descriptor().Fields().ByName(protoreflect.Name(signerField))
			if field == nil {
				return nil, fmt.Errorf("%s signer field %q not found", reflectedMsg.Descriptor().FullName(), signerField)
			}
			if field.Kind() != protoreflect.BytesKind {
				return nil, fmt.Errorf("%s signer field %q must be bytes", reflectedMsg.Descriptor().FullName(), signerField)
			}
			signer := reflectedMsg.Get(field).Bytes()
			return [][]byte{append([]byte(nil), signer...)}, nil
		})
	}
}

func accAddressSignersToBytes(signers []sdk.AccAddress) [][]byte {
	result := make([][]byte, 0, len(signers))
	for _, signer := range signers {
		result = append(result, append([]byte(nil), signer...))
	}
	return result
}
