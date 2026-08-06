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

	for msgType, signerFieldName := range msgSignerFields() {
		msgType := msgType
		signerFieldName := signerFieldName
		t.Run(msgType.Elem().Name(), func(t *testing.T) {
			descriptor, err := gogoproto.HybridResolver.FindDescriptorByName(protoreflect.FullName("dex." + msgType.Elem().Name()))
			require.NoError(t, err)
			msgDescriptor, ok := descriptor.(protoreflect.MessageDescriptor)
			require.True(t, ok)
			msg := dynamicpb.NewMessage(msgDescriptor)
			signerField := msg.Descriptor().Fields().ByName(protoreflect.Name(signerFieldName))
			require.NotNil(t, signerField)
			require.Equal(t, protoreflect.BytesKind, signerField.Kind())
			expected := []byte("twenty-byte-address!")
			msg.Set(signerField, protoreflect.ValueOfBytes(expected))

			getSigners := options.CustomGetSigners[msg.Descriptor().FullName()]
			require.NotNil(t, getSigners)
			signers, err := getSigners(msg)
			require.NoError(t, err)
			require.Equal(t, [][]byte{expected}, signers)

			// The resolver must return an owned copy rather than the dynamic
			// message's mutable backing storage.
			expected[0] = 'X'
			require.Equal(t, byte('t'), signers[0][0])
		})
	}
	require.Len(t, msgSignerFields(), len(msgTypesForDescriptor()))
}
