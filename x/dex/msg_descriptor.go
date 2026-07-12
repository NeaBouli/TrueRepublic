package dex

import (
	"bytes"
	"compress/gzip"

	gogoproto "github.com/cosmos/gogoproto/proto"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	msgDescriptorFile   = "dex/tx.proto"
	queryDescriptorFile = "dex/query.proto"
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
	registerCompressedMsgDescriptor(msgDescriptorFile, file)
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

func registerCompressedMsgDescriptor(name string, file *descriptorpb.FileDescriptorProto) {
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
	gogoproto.RegisterFile(name, compressed.Bytes())
}
