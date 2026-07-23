package truedemocracy

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
	msgDescriptorFile   = "truedemocracy/tx.proto"
	queryDescriptorFile = "truedemocracy/query.proto"
)

var (
	msgDescriptorBytes []byte
	msgDescriptorIndex = map[string][]int{}
)

func registerMsgFileDescriptor() {
	service := &descriptorpb.ServiceDescriptorProto{Name: proto2.String("Msg")}
	for _, method := range _Msg_serviceDesc.Methods {
		messageName := ".truedemocracy.Msg" + method.MethodName
		service.Method = append(service.Method, &descriptorpb.MethodDescriptorProto{
			Name: proto2.String(method.MethodName), InputType: proto2.String(messageName), OutputType: proto2.String(messageName + "Response"),
		})
	}
	file := &descriptorpb.FileDescriptorProto{
		Name:       proto2.String(msgDescriptorFile),
		Package:    proto2.String("truedemocracy"),
		Syntax:     proto2.String("proto3"),
		Dependency: []string{"cosmos/base/v1beta1/coin.proto"},
		Service:    []*descriptorpb.ServiceDescriptorProto{service},
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
		messageName := ".truedemocracy.Query" + method.MethodName
		service.Method = append(service.Method, &descriptorpb.MethodDescriptorProto{
			Name: proto2.String(method.MethodName), InputType: proto2.String(messageName + "Request"), OutputType: proto2.String(messageName + "Response"),
		})
	}
	file := &descriptorpb.FileDescriptorProto{
		Name: proto2.String(queryDescriptorFile), Package: proto2.String("truedemocracy"), Syntax: proto2.String("proto3"),
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
		reflect.TypeOf((*MsgCreateDomain)(nil)),
		reflect.TypeOf((*MsgSubmitProposal)(nil)),
		reflect.TypeOf((*MsgRegisterValidator)(nil)),
		reflect.TypeOf((*MsgWithdrawStake)(nil)),
		reflect.TypeOf((*MsgRemoveValidator)(nil)),
		reflect.TypeOf((*MsgRotateValidatorKey)(nil)),
		reflect.TypeOf((*MsgUnjail)(nil)),
		reflect.TypeOf((*MsgJoinPermissionRegister)(nil)),
		reflect.TypeOf((*MsgPurgePermissionRegister)(nil)),
		reflect.TypeOf((*MsgPlaceStoneOnIssue)(nil)),
		reflect.TypeOf((*MsgPlaceStoneOnSuggestion)(nil)),
		reflect.TypeOf((*MsgPlaceStoneOnMember)(nil)),
		reflect.TypeOf((*MsgVoteToExclude)(nil)),
		reflect.TypeOf((*MsgVoteToDelete)(nil)),
		reflect.TypeOf((*MsgRateProposal)(nil)),
		reflect.TypeOf((*MsgCastElectionVote)(nil)),
		reflect.TypeOf((*MsgAddMember)(nil)),
		reflect.TypeOf((*MsgOnboardToDomain)(nil)),
		reflect.TypeOf((*MsgApproveOnboarding)(nil)),
		reflect.TypeOf((*MsgRejectOnboarding)(nil)),
		reflect.TypeOf((*MsgRegisterIdentity)(nil)),
		reflect.TypeOf((*MsgRateWithProof)(nil)),
		reflect.TypeOf((*MsgDepositToDomain)(nil)),
		reflect.TypeOf((*MsgWithdrawFromDomain)(nil)),
	}
}

func msgResponseTypesForDescriptor() []string {
	return []string{
		"MsgCreateDomainResponse",
		"MsgSubmitProposalResponse",
		"MsgRegisterValidatorResponse",
		"MsgWithdrawStakeResponse",
		"MsgRemoveValidatorResponse",
		"MsgRotateValidatorKeyResponse",
		"MsgUnjailResponse",
		"MsgJoinPermissionRegisterResponse",
		"MsgPurgePermissionRegisterResponse",
		"MsgPlaceStoneOnIssueResponse",
		"MsgPlaceStoneOnSuggestionResponse",
		"MsgPlaceStoneOnMemberResponse",
		"MsgVoteToExcludeResponse",
		"MsgVoteToDeleteResponse",
		"MsgRateProposalResponse",
		"MsgCastElectionVoteResponse",
		"MsgAddMemberResponse",
		"MsgOnboardToDomainResponse",
		"MsgApproveOnboardingResponse",
		"MsgRejectOnboardingResponse",
		"MsgRegisterIdentityResponse",
		"MsgRateWithProofResponse",
		"MsgDepositToDomainResponse",
		"MsgWithdrawFromDomainResponse",
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
	fieldType, typeName := descriptorTypeForGoField(field.Type)
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto2.String(protoName),
		Number:   proto2.Int32(int32(number)),
		Label:    &label,
		Type:     &fieldType,
		TypeName: typeName,
	}
}

func descriptorTypeForGoField(fieldType reflect.Type) (descriptorpb.FieldDescriptorProto_Type, *string) {
	if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Uint8 {
		return descriptorpb.FieldDescriptorProto_TYPE_BYTES, nil
	}
	if fieldType.Kind() == reflect.Slice && fieldType.Elem().Name() == "Coin" {
		return descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, proto2.String(".cosmos.base.v1beta1.Coin")
	}
	switch fieldType.Kind() {
	case reflect.String:
		return descriptorpb.FieldDescriptorProto_TYPE_STRING, nil
	case reflect.Int32:
		return descriptorpb.FieldDescriptorProto_TYPE_INT32, nil
	case reflect.Int64:
		return descriptorpb.FieldDescriptorProto_TYPE_INT64, nil
	case reflect.Struct:
		if fieldType.Name() == "Coin" {
			return descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, proto2.String(".cosmos.base.v1beta1.Coin")
		}
	}
	panic("unsupported protobuf descriptor field type " + fieldType.String())
}

func descriptorForMessage(name string) ([]byte, []int) {
	return msgDescriptorBytes, msgDescriptorIndex[name]
}

func (*MsgCreateDomain) Descriptor() ([]byte, []int) { return descriptorForMessage("MsgCreateDomain") }
func (*MsgSubmitProposal) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgSubmitProposal")
}
func (*MsgRegisterValidator) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRegisterValidator")
}
func (*MsgWithdrawStake) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgWithdrawStake")
}
func (*MsgRemoveValidator) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRemoveValidator")
}
func (*MsgRotateValidatorKey) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRotateValidatorKey")
}
func (*MsgUnjail) Descriptor() ([]byte, []int) { return descriptorForMessage("MsgUnjail") }
func (*MsgJoinPermissionRegister) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgJoinPermissionRegister")
}
func (*MsgPurgePermissionRegister) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgPurgePermissionRegister")
}
func (*MsgPlaceStoneOnIssue) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgPlaceStoneOnIssue")
}
func (*MsgPlaceStoneOnSuggestion) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgPlaceStoneOnSuggestion")
}
func (*MsgPlaceStoneOnMember) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgPlaceStoneOnMember")
}
func (*MsgVoteToExclude) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgVoteToExclude")
}
func (*MsgVoteToDelete) Descriptor() ([]byte, []int) { return descriptorForMessage("MsgVoteToDelete") }
func (*MsgRateProposal) Descriptor() ([]byte, []int) { return descriptorForMessage("MsgRateProposal") }
func (*MsgCastElectionVote) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgCastElectionVote")
}
func (*MsgAddMember) Descriptor() ([]byte, []int) { return descriptorForMessage("MsgAddMember") }
func (*MsgOnboardToDomain) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgOnboardToDomain")
}
func (*MsgApproveOnboarding) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgApproveOnboarding")
}
func (*MsgRejectOnboarding) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRejectOnboarding")
}
func (*MsgRegisterIdentity) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRegisterIdentity")
}
func (*MsgRateWithProof) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRateWithProof")
}
func (*MsgDepositToDomain) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgDepositToDomain")
}
func (*MsgWithdrawFromDomain) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgWithdrawFromDomain")
}
func (*MsgCreateDomainResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgCreateDomainResponse")
}
func (*MsgSubmitProposalResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgSubmitProposalResponse")
}
func (*MsgRegisterValidatorResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRegisterValidatorResponse")
}
func (*MsgWithdrawStakeResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgWithdrawStakeResponse")
}
func (*MsgRemoveValidatorResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRemoveValidatorResponse")
}
func (*MsgRotateValidatorKeyResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRotateValidatorKeyResponse")
}
func (*MsgUnjailResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgUnjailResponse")
}
func (*MsgJoinPermissionRegisterResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgJoinPermissionRegisterResponse")
}
func (*MsgPurgePermissionRegisterResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgPurgePermissionRegisterResponse")
}
func (*MsgPlaceStoneOnIssueResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgPlaceStoneOnIssueResponse")
}
func (*MsgPlaceStoneOnSuggestionResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgPlaceStoneOnSuggestionResponse")
}
func (*MsgPlaceStoneOnMemberResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgPlaceStoneOnMemberResponse")
}
func (*MsgVoteToExcludeResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgVoteToExcludeResponse")
}
func (*MsgVoteToDeleteResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgVoteToDeleteResponse")
}
func (*MsgRateProposalResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRateProposalResponse")
}
func (*MsgCastElectionVoteResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgCastElectionVoteResponse")
}
func (*MsgAddMemberResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgAddMemberResponse")
}
func (*MsgOnboardToDomainResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgOnboardToDomainResponse")
}
func (*MsgApproveOnboardingResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgApproveOnboardingResponse")
}
func (*MsgRejectOnboardingResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRejectOnboardingResponse")
}
func (*MsgRegisterIdentityResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRegisterIdentityResponse")
}
func (*MsgRateWithProofResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgRateWithProofResponse")
}
func (*MsgDepositToDomainResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgDepositToDomainResponse")
}
func (*MsgWithdrawFromDomainResponse) Descriptor() ([]byte, []int) {
	return descriptorForMessage("MsgWithdrawFromDomainResponse")
}
