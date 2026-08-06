package dex

import (
	"bytes"
	"compress/gzip"
	"reflect"
	"strconv"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	msgDescriptorFile   = "dex/tx.proto"
	queryDescriptorFile = "dex/query.proto"
)

var (
	msgDescriptorBytes []byte
	msgDescriptorIndex = map[string][]int{}
)

func registerMsgFileDescriptor() {
	service := &descriptorpb.ServiceDescriptorProto{Name: proto2.String("Msg")}
	for _, method := range _Msg_serviceDesc.Methods {
		messageName := ".dex.Msg" + method.MethodName
		service.Method = append(service.Method, &descriptorpb.MethodDescriptorProto{
			Name: proto2.String(method.MethodName), InputType: proto2.String(messageName), OutputType: proto2.String(messageName + "Response"),
		})
	}
	file := &descriptorpb.FileDescriptorProto{
		Name: proto2.String(msgDescriptorFile), Package: proto2.String("dex"), Syntax: proto2.String("proto3"),
		Service: []*descriptorpb.ServiceDescriptorProto{service},
	}
	for _, typ := range msgTypesForDescriptor() {
		msgDescriptorIndex[typ.Elem().Name()] = []int{len(file.MessageType)}
		file.MessageType = append(file.MessageType, buildMessageDescriptor(typ))
	}
	for _, name := range msgResponseTypesForDescriptor() {
		msgDescriptorIndex[name] = []int{len(file.MessageType)}
		file.MessageType = append(file.MessageType, &descriptorpb.DescriptorProto{Name: proto2.String(name)})
	}
	msgDescriptorBytes = registerCompressedMsgDescriptor(msgDescriptorFile, file)
}

func registerQueryFileDescriptor() {
	service := &descriptorpb.ServiceDescriptorProto{Name: proto2.String("Query")}
	for _, method := range _Query_serviceDesc.Methods {
		messageName := ".dex.Query" + method.MethodName
		service.Method = append(service.Method, &descriptorpb.MethodDescriptorProto{
			Name: proto2.String(method.MethodName), InputType: proto2.String(messageName + "Request"), OutputType: proto2.String(messageName + "Response"),
		})
	}
	file := &descriptorpb.FileDescriptorProto{
		Name: proto2.String(queryDescriptorFile), Package: proto2.String("dex"), Syntax: proto2.String("proto3"),
		Service: []*descriptorpb.ServiceDescriptorProto{service},
	}
	registerCompressedMsgDescriptor(queryDescriptorFile, file)
}

func registerCompressedMsgDescriptor(name string, file *descriptorpb.FileDescriptorProto) []byte {
	raw, err := proto2.Marshal(file)
	if err != nil {
		panic(err)
	}
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(raw); err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}
	bz := compressed.Bytes()
	gogoproto.RegisterFile(name, bz)
	return bz
}

func msgTypesForDescriptor() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*MsgCreatePool)(nil)),
		reflect.TypeOf((*MsgSwap)(nil)),
		reflect.TypeOf((*MsgAddLiquidity)(nil)),
		reflect.TypeOf((*MsgRemoveLiquidity)(nil)),
		reflect.TypeOf((*MsgRegisterAsset)(nil)),
		reflect.TypeOf((*MsgUpdateAssetStatus)(nil)),
		reflect.TypeOf((*MsgSwapExact)(nil)),
	}
}

// msgSignerFields is the explicit signing contract for every hand-written DEX
// message registered with the SDK. Keeping it separate from descriptor
// discovery prevents a descriptor-only message from silently inheriting a
// hard-coded signer assumption.
func msgSignerFields() map[reflect.Type]string {
	return map[reflect.Type]string{
		reflect.TypeOf((*MsgCreatePool)(nil)):        "sender",
		reflect.TypeOf((*MsgSwap)(nil)):              "sender",
		reflect.TypeOf((*MsgAddLiquidity)(nil)):      "sender",
		reflect.TypeOf((*MsgRemoveLiquidity)(nil)):   "sender",
		reflect.TypeOf((*MsgRegisterAsset)(nil)):     "sender",
		reflect.TypeOf((*MsgUpdateAssetStatus)(nil)): "sender",
		reflect.TypeOf((*MsgSwapExact)(nil)):         "sender",
	}
}

func msgResponseTypesForDescriptor() []string {
	return []string{
		"MsgCreatePoolResponse",
		"MsgSwapResponse",
		"MsgAddLiquidityResponse",
		"MsgRemoveLiquidityResponse",
		"MsgRegisterAssetResponse",
		"MsgUpdateAssetStatusResponse",
		"MsgSwapExactResponse",
	}
}

func buildMessageDescriptor(pointerType reflect.Type) *descriptorpb.DescriptorProto {
	structType := pointerType.Elem()
	desc := &descriptorpb.DescriptorProto{Name: proto2.String(structType.Name())}
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("protobuf")
		if tag == "" {
			continue
		}
		desc.Field = append(desc.Field, buildFieldDescriptor(field, tag))
	}
	return desc
}

func buildFieldDescriptor(field reflect.StructField, tag string) *descriptorpb.FieldDescriptorProto {
	parts := strings.Split(tag, ",")
	if len(parts) < 3 {
		panic("invalid protobuf tag for " + field.Name)
	}
	number, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err)
	}
	protoName := ""
	repeated := false
	for _, part := range parts[2:] {
		switch {
		case part == "rep":
			repeated = true
		case strings.HasPrefix(part, "name="):
			protoName = strings.TrimPrefix(part, "name=")
		}
	}
	if protoName == "" {
		panic("missing protobuf name for " + field.Name)
	}
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	if repeated {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	}
	fieldType := descriptorTypeForGoField(field.Type)
	expectedWireType := "varint"
	if fieldType == descriptorpb.FieldDescriptorProto_TYPE_BYTES || fieldType == descriptorpb.FieldDescriptorProto_TYPE_STRING {
		expectedWireType = "bytes"
	}
	if parts[0] != expectedWireType {
		panic("protobuf wire type " + parts[0] + " does not match " + expectedWireType + " for " + field.Name)
	}
	return &descriptorpb.FieldDescriptorProto{
		Name: proto2.String(protoName), Number: proto2.Int32(int32(number)), Label: &label, Type: &fieldType,
	}
}

func descriptorTypeForGoField(fieldType reflect.Type) descriptorpb.FieldDescriptorProto_Type {
	if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Uint8 {
		return descriptorpb.FieldDescriptorProto_TYPE_BYTES
	}
	switch fieldType.Kind() {
	case reflect.String:
		return descriptorpb.FieldDescriptorProto_TYPE_STRING
	case reflect.Int64:
		return descriptorpb.FieldDescriptorProto_TYPE_INT64
	case reflect.Uint32:
		return descriptorpb.FieldDescriptorProto_TYPE_UINT32
	case reflect.Bool:
		return descriptorpb.FieldDescriptorProto_TYPE_BOOL
	default:
		panic("unsupported protobuf descriptor field type " + fieldType.String())
	}
}

func descriptorForMessage(name string) ([]byte, []int) {
	return msgDescriptorBytes, msgDescriptorIndex[name]
}

func (*MsgCreatePool) Descriptor() ([]byte, []int) { return descriptorForMessage("MsgCreatePool") }
func (*MsgSwap) Descriptor() ([]byte, []int)       { return descriptorForMessage("MsgSwap") }
func (*MsgAddLiquidity) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgAddLiquidity")
}
func (*MsgRemoveLiquidity) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRemoveLiquidity")
}
func (*MsgRegisterAsset) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRegisterAsset")
}
func (*MsgUpdateAssetStatus) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgUpdateAssetStatus")
}
func (*MsgSwapExact) Descriptor() ([]byte, []int) { return descriptorForMessage("MsgSwapExact") }
func (*MsgCreatePoolResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgCreatePoolResponse")
}
func (*MsgSwapResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgSwapResponse")
}
func (*MsgAddLiquidityResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgAddLiquidityResponse")
}
func (*MsgRemoveLiquidityResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRemoveLiquidityResponse")
}
func (*MsgRegisterAssetResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRegisterAssetResponse")
}
func (*MsgUpdateAssetStatusResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgUpdateAssetStatusResponse")
}
func (*MsgSwapExactResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgSwapExactResponse")
}
