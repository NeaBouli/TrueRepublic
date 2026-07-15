package truedemocracy

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

var msgSignerFields = map[string]protoreflect.Name{
	"MsgCreateDomain":            "admin",
	"MsgPurgePermissionRegister": "caller",
}

// RegisterCustomGetSigners wires the hand-written truedemocracy message types
// into the Cosmos SDK v0.50 signing context. The messages already expose the
// legacy sdk.Msg GetSigners contract; this adapter lets x/tx resolve signers
// without generated cosmos.msg.v1.signer protobuf annotations.
func RegisterCustomGetSigners(options *txsigning.Options) {
	for _, msgType := range msgTypesForDescriptor() {
		msgName := msgType.Elem().Name()
		signerField := msgSignerFields[msgName]
		if signerField == "" {
			signerField = "sender"
		}

		typeName := protoreflect.FullName("truedemocracy." + msgName)
		options.DefineCustomGetSigners(typeName, func(msg proto.Message) ([][]byte, error) {
			return getMsgSigners(msg, signerField)
		})
	}
}

func getMsgSigners(msg proto.Message, signerField protoreflect.Name) ([][]byte, error) {
	if legacyMsg, ok := msg.(legacyMsgSigner); ok {
		return accAddressSignersToBytes(legacyMsg.GetSigners()), nil
	}

	reflectedMsg := msg.ProtoReflect()
	field := reflectedMsg.Descriptor().Fields().ByName(signerField)
	if field == nil {
		return nil, fmt.Errorf("%s signer field %q not found", reflectedMsg.Descriptor().FullName(), signerField)
	}
	if field.Kind() != protoreflect.BytesKind {
		return nil, fmt.Errorf("%s signer field %q must be bytes", reflectedMsg.Descriptor().FullName(), signerField)
	}

	signer := reflectedMsg.Get(field).Bytes()
	if len(signer) == 0 {
		return [][]byte{nil}, nil
	}
	signerCopy := make([]byte, len(signer))
	copy(signerCopy, signer)
	return [][]byte{signerCopy}, nil
}

func accAddressSignersToBytes(signers []sdk.AccAddress) [][]byte {
	signerBytes := make([][]byte, 0, len(signers))
	for _, signer := range signers {
		if signer == nil {
			signerBytes = append(signerBytes, nil)
			continue
		}
		signerCopy := make([]byte, len(signer))
		copy(signerCopy, signer)
		signerBytes = append(signerBytes, signerCopy)
	}
	return signerBytes
}
