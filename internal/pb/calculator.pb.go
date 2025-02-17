// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: calculator.proto

package pb

import (
	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type FibonacciOfRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// first declares the first number of the sequence.
	First int64 `protobuf:"varint,1,opt,name=first,proto3" json:"first,omitempty"`
	// second declares the second number of the sequence.
	Second int64 `protobuf:"varint,2,opt,name=second,proto3" json:"second,omitempty"`
	// nth_position defines the Nth position of the sequence to calculate the
	// number of. The first number in the sequence is at position 1.
	NthPosition int64 `protobuf:"varint,3,opt,name=nth_position,json=nthPosition,proto3" json:"nth_position,omitempty"`
}

func (x *FibonacciOfRequest) Reset() {
	*x = FibonacciOfRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calculator_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FibonacciOfRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FibonacciOfRequest) ProtoMessage() {}

func (x *FibonacciOfRequest) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FibonacciOfRequest.ProtoReflect.Descriptor instead.
func (*FibonacciOfRequest) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{0}
}

func (x *FibonacciOfRequest) GetFirst() int64 {
	if x != nil {
		return x.First
	}
	return 0
}

func (x *FibonacciOfRequest) GetSecond() int64 {
	if x != nil {
		return x.Second
	}
	return 0
}

func (x *FibonacciOfRequest) GetNthPosition() int64 {
	if x != nil {
		return x.NthPosition
	}
	return 0
}

type FibonacciOfResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// first declares the first number of the sequence.
	First int64 `protobuf:"varint,1,opt,name=first,proto3" json:"first,omitempty"`
	// second declares the second number of the sequence.
	Second int64 `protobuf:"varint,2,opt,name=second,proto3" json:"second,omitempty"`
	// nth_position indexes the calculated fibonacci number at this position
	// starting at 1.
	NthPosition int64 `protobuf:"varint,3,opt,name=nth_position,json=nthPosition,proto3" json:"nth_position,omitempty"`
	// result is the calculated fibonacci number at the requested position.
	Result int64 `protobuf:"varint,4,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *FibonacciOfResponse) Reset() {
	*x = FibonacciOfResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calculator_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FibonacciOfResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FibonacciOfResponse) ProtoMessage() {}

func (x *FibonacciOfResponse) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FibonacciOfResponse.ProtoReflect.Descriptor instead.
func (*FibonacciOfResponse) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{1}
}

func (x *FibonacciOfResponse) GetFirst() int64 {
	if x != nil {
		return x.First
	}
	return 0
}

func (x *FibonacciOfResponse) GetSecond() int64 {
	if x != nil {
		return x.Second
	}
	return 0
}

func (x *FibonacciOfResponse) GetNthPosition() int64 {
	if x != nil {
		return x.NthPosition
	}
	return 0
}

func (x *FibonacciOfResponse) GetResult() int64 {
	if x != nil {
		return x.Result
	}
	return 0
}

// TODO. Either take inspiration from google.api.servicemanagement.v1 or reuse
// it.
type CalculationMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Created *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=created,proto3" json:"created,omitempty"`
	Started *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=started,proto3" json:"started,omitempty"`
}

func (x *CalculationMetadata) Reset() {
	*x = CalculationMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calculator_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CalculationMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CalculationMetadata) ProtoMessage() {}

func (x *CalculationMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CalculationMetadata.ProtoReflect.Descriptor instead.
func (*CalculationMetadata) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{2}
}

func (x *CalculationMetadata) GetCreated() *timestamppb.Timestamp {
	if x != nil {
		return x.Created
	}
	return nil
}

func (x *CalculationMetadata) GetStarted() *timestamppb.Timestamp {
	if x != nil {
		return x.Started
	}
	return nil
}

var File_calculator_proto protoreflect.FileDescriptor

var file_calculator_proto_rawDesc = []byte{
	0x0a, 0x10, 0x63, 0x61, 0x6c, 0x63, 0x75, 0x6c, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0a, 0x63, 0x61, 0x6c, 0x63, 0x75, 0x6c, 0x61, 0x74, 0x6f, 0x72, 0x1a, 0x23,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x6c, 0x6f, 0x6e, 0x67, 0x72, 0x75, 0x6e, 0x6e, 0x69,
	0x6e, 0x67, 0x2f, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x65, 0x0a, 0x12, 0x46, 0x69, 0x62, 0x6f, 0x6e, 0x61, 0x63, 0x63,
	0x69, 0x4f, 0x66, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x69,
	0x72, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x66, 0x69, 0x72, 0x73, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x06, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x6e, 0x74, 0x68, 0x5f,
	0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b,
	0x6e, 0x74, 0x68, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x7e, 0x0a, 0x13, 0x46,
	0x69, 0x62, 0x6f, 0x6e, 0x61, 0x63, 0x63, 0x69, 0x4f, 0x66, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x69, 0x72, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x05, 0x66, 0x69, 0x72, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x63, 0x6f,
	0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64,
	0x12, 0x21, 0x0a, 0x0c, 0x6e, 0x74, 0x68, 0x5f, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x6e, 0x74, 0x68, 0x50, 0x6f, 0x73, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x81, 0x01, 0x0a, 0x13,
	0x43, 0x61, 0x6c, 0x63, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x34, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x34, 0x0a, 0x07, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x32,
	0xac, 0x03, 0x0a, 0x0c, 0x43, 0x61, 0x6c, 0x63, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x7b, 0x0a, 0x0b, 0x46, 0x69, 0x62, 0x6f, 0x6e, 0x61, 0x63, 0x63, 0x69, 0x4f, 0x66, 0x12,
	0x1e, 0x2e, 0x63, 0x61, 0x6c, 0x63, 0x75, 0x6c, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x46, 0x69, 0x62,
	0x6f, 0x6e, 0x61, 0x63, 0x63, 0x69, 0x4f, 0x66, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x6e, 0x67, 0x72, 0x75, 0x6e,
	0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x2d,
	0xca, 0x41, 0x2a, 0x0a, 0x13, 0x46, 0x69, 0x62, 0x6f, 0x6e, 0x61, 0x63, 0x63, 0x69, 0x4f, 0x66,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x13, 0x43, 0x61, 0x6c, 0x63, 0x75, 0x6c,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x58, 0x0a,
	0x0c, 0x47, 0x65, 0x74, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x27, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x6e, 0x67, 0x72, 0x75, 0x6e, 0x6e, 0x69,
	0x6e, 0x67, 0x2e, 0x47, 0x65, 0x74, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x6c, 0x6f, 0x6e, 0x67, 0x72, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x4f, 0x70, 0x65, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x00, 0x12, 0x69, 0x0a, 0x0e, 0x4c, 0x69, 0x73, 0x74, 0x4f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x29, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x6e, 0x67, 0x72, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x4c,
	0x69, 0x73, 0x74, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x6c, 0x6f,
	0x6e, 0x67, 0x72, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4f, 0x70,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x5a, 0x0a, 0x0d, 0x57, 0x61, 0x69, 0x74, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x28, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x6e,
	0x67, 0x72, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x57, 0x61, 0x69, 0x74, 0x4f, 0x70, 0x65,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x6e, 0x67, 0x72, 0x75, 0x6e, 0x6e, 0x69,
	0x6e, 0x67, 0x2e, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x00, 0x42, 0x2e,
	0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x69, 0x63,
	0x6b, 0x6c, 0x65, 0x66, 0x6f, 0x72, 0x64, 0x2f, 0x63, 0x61, 0x6c, 0x63, 0x75, 0x6c, 0x61, 0x74,
	0x6f, 0x72, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_calculator_proto_rawDescOnce sync.Once
	file_calculator_proto_rawDescData = file_calculator_proto_rawDesc
)

func file_calculator_proto_rawDescGZIP() []byte {
	file_calculator_proto_rawDescOnce.Do(func() {
		file_calculator_proto_rawDescData = protoimpl.X.CompressGZIP(file_calculator_proto_rawDescData)
	})
	return file_calculator_proto_rawDescData
}

var file_calculator_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_calculator_proto_goTypes = []any{
	(*FibonacciOfRequest)(nil),                   // 0: calculator.FibonacciOfRequest
	(*FibonacciOfResponse)(nil),                  // 1: calculator.FibonacciOfResponse
	(*CalculationMetadata)(nil),                  // 2: calculator.CalculationMetadata
	(*timestamppb.Timestamp)(nil),                // 3: google.protobuf.Timestamp
	(*longrunningpb.GetOperationRequest)(nil),    // 4: google.longrunning.GetOperationRequest
	(*longrunningpb.ListOperationsRequest)(nil),  // 5: google.longrunning.ListOperationsRequest
	(*longrunningpb.WaitOperationRequest)(nil),   // 6: google.longrunning.WaitOperationRequest
	(*longrunningpb.Operation)(nil),              // 7: google.longrunning.Operation
	(*longrunningpb.ListOperationsResponse)(nil), // 8: google.longrunning.ListOperationsResponse
}
var file_calculator_proto_depIdxs = []int32{
	3, // 0: calculator.CalculationMetadata.created:type_name -> google.protobuf.Timestamp
	3, // 1: calculator.CalculationMetadata.started:type_name -> google.protobuf.Timestamp
	0, // 2: calculator.Calculations.FibonacciOf:input_type -> calculator.FibonacciOfRequest
	4, // 3: calculator.Calculations.GetOperation:input_type -> google.longrunning.GetOperationRequest
	5, // 4: calculator.Calculations.ListOperations:input_type -> google.longrunning.ListOperationsRequest
	6, // 5: calculator.Calculations.WaitOperation:input_type -> google.longrunning.WaitOperationRequest
	7, // 6: calculator.Calculations.FibonacciOf:output_type -> google.longrunning.Operation
	7, // 7: calculator.Calculations.GetOperation:output_type -> google.longrunning.Operation
	8, // 8: calculator.Calculations.ListOperations:output_type -> google.longrunning.ListOperationsResponse
	7, // 9: calculator.Calculations.WaitOperation:output_type -> google.longrunning.Operation
	6, // [6:10] is the sub-list for method output_type
	2, // [2:6] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_calculator_proto_init() }
func file_calculator_proto_init() {
	if File_calculator_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_calculator_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*FibonacciOfRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_calculator_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*FibonacciOfResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_calculator_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*CalculationMetadata); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_calculator_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_calculator_proto_goTypes,
		DependencyIndexes: file_calculator_proto_depIdxs,
		MessageInfos:      file_calculator_proto_msgTypes,
	}.Build()
	File_calculator_proto = out.File
	file_calculator_proto_rawDesc = nil
	file_calculator_proto_goTypes = nil
	file_calculator_proto_depIdxs = nil
}
