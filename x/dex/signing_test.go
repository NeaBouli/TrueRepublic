package dex

import (
	"testing"

	txsigning "cosmossdk.io/x/tx/signing"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestCustomGetSignersSupportsDynamicMessages(t *testing.T) {
	options := &txsigning.Options{}
	RegisterCustomGetSigners(options)

	descriptor, err := gogoproto.HybridResolver.FindDescriptorByName("dex.MsgCreatePool")
	require.NoError(t, err)
	msgDescriptor, ok := descriptor.(protoreflect.MessageDescriptor)
	require.True(t, ok)
	msg := dynamicpb.NewMessage(msgDescriptor)
	senderField := msg.Descriptor().Fields().ByName("sender")
	require.NotNil(t, senderField)
	expected := []byte("twenty-byte-address!")
	msg.Set(senderField, protoreflect.ValueOfBytes(expected))

	getSigners := options.CustomGetSigners[msg.Descriptor().FullName()]
	require.NotNil(t, getSigners)
	signers, err := getSigners(msg)
	require.NoError(t, err)
	require.Equal(t, [][]byte{expected}, signers)

	// The resolver must return an owned copy rather than the dynamic message's
	// mutable backing storage.
	expected[0] = 'X'
	require.Equal(t, byte('t'), signers[0][0])
}
