// Code generated by protoc-gen-go. DO NOT EDIT.
// source: payment-gateway-hook.proto

package payment_gateway

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type PaygateHookRequest struct {
	OrderID              string   `protobuf:"bytes,1,opt,name=orderID,proto3" json:"orderID,omitempty"`
	PaymentId            string   `protobuf:"bytes,2,opt,name=paymentId,proto3" json:"paymentId,omitempty"`
	InvoiceId            int64    `protobuf:"varint,3,opt,name=invoiceId,proto3" json:"invoiceId,omitempty"`
	Amount               int64    `protobuf:"varint,4,opt,name=amount,proto3" json:"amount,omitempty"`
	ReqBody              string   `protobuf:"bytes,5,opt,name=reqBody,proto3" json:"reqBody,omitempty"`
	ResBody              string   `protobuf:"bytes,6,opt,name=resBody,proto3" json:"resBody,omitempty"`
	CardMask             string   `protobuf:"bytes,7,opt,name=cardMask,proto3" json:"cardMask,omitempty"`
	Result               bool     `protobuf:"varint,8,opt,name=result,proto3" json:"result,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PaygateHookRequest) Reset()         { *m = PaygateHookRequest{} }
func (m *PaygateHookRequest) String() string { return proto.CompactTextString(m) }
func (*PaygateHookRequest) ProtoMessage()    {}
func (*PaygateHookRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_afe8d80d249138bb, []int{0}
}

func (m *PaygateHookRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PaygateHookRequest.Unmarshal(m, b)
}
func (m *PaygateHookRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PaygateHookRequest.Marshal(b, m, deterministic)
}
func (m *PaygateHookRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PaygateHookRequest.Merge(m, src)
}
func (m *PaygateHookRequest) XXX_Size() int {
	return xxx_messageInfo_PaygateHookRequest.Size(m)
}
func (m *PaygateHookRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PaygateHookRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PaygateHookRequest proto.InternalMessageInfo

func (m *PaygateHookRequest) GetOrderID() string {
	if m != nil {
		return m.OrderID
	}
	return ""
}

func (m *PaygateHookRequest) GetPaymentId() string {
	if m != nil {
		return m.PaymentId
	}
	return ""
}

func (m *PaygateHookRequest) GetInvoiceId() int64 {
	if m != nil {
		return m.InvoiceId
	}
	return 0
}

func (m *PaygateHookRequest) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *PaygateHookRequest) GetReqBody() string {
	if m != nil {
		return m.ReqBody
	}
	return ""
}

func (m *PaygateHookRequest) GetResBody() string {
	if m != nil {
		return m.ResBody
	}
	return ""
}

func (m *PaygateHookRequest) GetCardMask() string {
	if m != nil {
		return m.CardMask
	}
	return ""
}

func (m *PaygateHookRequest) GetResult() bool {
	if m != nil {
		return m.Result
	}
	return false
}

type PaygateHookResponse struct {
	Ok                   bool     `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PaygateHookResponse) Reset()         { *m = PaygateHookResponse{} }
func (m *PaygateHookResponse) String() string { return proto.CompactTextString(m) }
func (*PaygateHookResponse) ProtoMessage()    {}
func (*PaygateHookResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_afe8d80d249138bb, []int{1}
}

func (m *PaygateHookResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PaygateHookResponse.Unmarshal(m, b)
}
func (m *PaygateHookResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PaygateHookResponse.Marshal(b, m, deterministic)
}
func (m *PaygateHookResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PaygateHookResponse.Merge(m, src)
}
func (m *PaygateHookResponse) XXX_Size() int {
	return xxx_messageInfo_PaygateHookResponse.Size(m)
}
func (m *PaygateHookResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PaygateHookResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PaygateHookResponse proto.InternalMessageInfo

func (m *PaygateHookResponse) GetOk() bool {
	if m != nil {
		return m.Ok
	}
	return false
}

func init() {
	proto.RegisterType((*PaygateHookRequest)(nil), "payment_gateway.PaygateHookRequest")
	proto.RegisterType((*PaygateHookResponse)(nil), "payment_gateway.PaygateHookResponse")
}

func init() { proto.RegisterFile("payment-gateway-hook.proto", fileDescriptor_afe8d80d249138bb) }

var fileDescriptor_afe8d80d249138bb = []byte{
	// 270 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xd1, 0x4a, 0xc3, 0x30,
	0x14, 0x86, 0x6d, 0xa7, 0x5d, 0x97, 0x8b, 0x09, 0x11, 0x24, 0x14, 0x2f, 0x4a, 0x55, 0xe8, 0xcd,
	0x7a, 0xa1, 0x6f, 0x30, 0x04, 0xed, 0x85, 0x20, 0x79, 0x01, 0x89, 0x4b, 0xd0, 0x91, 0x2d, 0xa7,
	0x4b, 0x52, 0xa5, 0xaf, 0xec, 0x53, 0x48, 0x4e, 0x33, 0x87, 0x0a, 0x5e, 0x7e, 0xe7, 0x3b, 0xf4,
	0xf4, 0xff, 0x43, 0x8a, 0x4e, 0x0c, 0x5b, 0x65, 0xfc, 0xe2, 0x55, 0x78, 0xf5, 0x21, 0x86, 0xc5,
	0x1b, 0x80, 0x6e, 0x3a, 0x0b, 0x1e, 0xe8, 0x69, 0x74, 0xcf, 0xd1, 0x55, 0x9f, 0x09, 0xa1, 0x4f,
	0x62, 0x08, 0xf8, 0x00, 0xa0, 0xb9, 0xda, 0xf5, 0xca, 0x79, 0xca, 0xc8, 0x14, 0xac, 0x54, 0xb6,
	0xbd, 0x63, 0x49, 0x99, 0xd4, 0x33, 0xbe, 0x47, 0x7a, 0x41, 0x66, 0xf1, 0x1b, 0xad, 0x64, 0x29,
	0xba, 0xc3, 0x20, 0xd8, 0xb5, 0x79, 0x87, 0xf5, 0x4a, 0xb5, 0x92, 0x4d, 0xca, 0xa4, 0x9e, 0xf0,
	0xc3, 0x80, 0x9e, 0x93, 0x4c, 0x6c, 0xa1, 0x37, 0x9e, 0x1d, 0xa3, 0x8a, 0x14, 0xae, 0x59, 0xb5,
	0x5b, 0x82, 0x1c, 0xd8, 0xc9, 0x78, 0x2d, 0xe2, 0x68, 0x1c, 0x9a, 0x6c, 0x6f, 0x10, 0x69, 0x41,
	0xf2, 0x95, 0xb0, 0xf2, 0x51, 0x38, 0xcd, 0xa6, 0xa8, 0xbe, 0x39, 0xdc, 0xb1, 0xca, 0xf5, 0x1b,
	0xcf, 0xf2, 0x32, 0xa9, 0x73, 0x1e, 0xa9, 0xba, 0x26, 0x67, 0x3f, 0xb2, 0xba, 0x0e, 0x8c, 0x53,
	0x74, 0x4e, 0x52, 0xd0, 0x98, 0x33, 0xe7, 0x29, 0xe8, 0x1b, 0x47, 0xe6, 0x4b, 0x61, 0x82, 0xef,
	0x37, 0x3e, 0x6c, 0x52, 0x81, 0x25, 0x85, 0x8c, 0xf7, 0x63, 0x6f, 0x38, 0xbd, 0x6c, 0x7e, 0xb5,
	0xd9, 0xfc, 0x6d, 0xb2, 0xb8, 0xfa, 0x7f, 0x69, 0xfc, 0x85, 0xea, 0xe8, 0x25, 0xc3, 0x07, 0xba,
	0xfd, 0x0a, 0x00, 0x00, 0xff, 0xff, 0xb7, 0x25, 0x98, 0xcb, 0xbe, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// BankResultHookClient is the client API for BankResultHook service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BankResultHookClient interface {
	PaymentGatewayHook(ctx context.Context, in *PaygateHookRequest, opts ...grpc.CallOption) (*PaygateHookResponse, error)
}

type bankResultHookClient struct {
	cc *grpc.ClientConn
}

func NewBankResultHookClient(cc *grpc.ClientConn) BankResultHookClient {
	return &bankResultHookClient{cc}
}

func (c *bankResultHookClient) PaymentGatewayHook(ctx context.Context, in *PaygateHookRequest, opts ...grpc.CallOption) (*PaygateHookResponse, error) {
	out := new(PaygateHookResponse)
	err := c.cc.Invoke(ctx, "/payment_gateway.BankResultHook/PaymentGatewayHook", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BankResultHookServer is the server API for BankResultHook service.
type BankResultHookServer interface {
	PaymentGatewayHook(context.Context, *PaygateHookRequest) (*PaygateHookResponse, error)
}

// UnimplementedBankResultHookServer can be embedded to have forward compatible implementations.
type UnimplementedBankResultHookServer struct {
}

func (*UnimplementedBankResultHookServer) PaymentGatewayHook(ctx context.Context, req *PaygateHookRequest) (*PaygateHookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PaymentGatewayHook not implemented")
}

func RegisterBankResultHookServer(s *grpc.Server, srv BankResultHookServer) {
	s.RegisterService(&_BankResultHook_serviceDesc, srv)
}

func _BankResultHook_PaymentGatewayHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PaygateHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BankResultHookServer).PaymentGatewayHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment_gateway.BankResultHook/PaymentGatewayHook",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BankResultHookServer).PaymentGatewayHook(ctx, req.(*PaygateHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _BankResultHook_serviceDesc = grpc.ServiceDesc{
	ServiceName: "payment_gateway.BankResultHook",
	HandlerType: (*BankResultHookServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PaymentGatewayHook",
			Handler:    _BankResultHook_PaymentGatewayHook_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment-gateway-hook.proto",
}
