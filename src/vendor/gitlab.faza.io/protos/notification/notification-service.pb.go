// Code generated by protoc-gen-go. DO NOT EDIT.
// source: notification-service.proto

package NotificationService

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

type OTP struct {
	Receiver             string   `protobuf:"bytes,1,opt,name=receiver,proto3" json:"receiver,omitempty"`
	Code                 string   `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OTP) Reset()         { *m = OTP{} }
func (m *OTP) String() string { return proto.CompactTextString(m) }
func (*OTP) ProtoMessage()    {}
func (*OTP) Descriptor() ([]byte, []int) {
	return fileDescriptor_62e234be764a7e86, []int{0}
}

func (m *OTP) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OTP.Unmarshal(m, b)
}
func (m *OTP) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OTP.Marshal(b, m, deterministic)
}
func (m *OTP) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OTP.Merge(m, src)
}
func (m *OTP) XXX_Size() int {
	return xxx_messageInfo_OTP.Size(m)
}
func (m *OTP) XXX_DiscardUnknown() {
	xxx_messageInfo_OTP.DiscardUnknown(m)
}

var xxx_messageInfo_OTP proto.InternalMessageInfo

func (m *OTP) GetReceiver() string {
	if m != nil {
		return m.Receiver
	}
	return ""
}

func (m *OTP) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

type EmailTemplate struct {
	TemplateName         string            `protobuf:"bytes,1,opt,name=templateName,proto3" json:"templateName,omitempty"`
	Vars                 map[string]string `protobuf:"bytes,2,rep,name=vars,proto3" json:"vars,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Email                *Email            `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *EmailTemplate) Reset()         { *m = EmailTemplate{} }
func (m *EmailTemplate) String() string { return proto.CompactTextString(m) }
func (*EmailTemplate) ProtoMessage()    {}
func (*EmailTemplate) Descriptor() ([]byte, []int) {
	return fileDescriptor_62e234be764a7e86, []int{1}
}

func (m *EmailTemplate) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EmailTemplate.Unmarshal(m, b)
}
func (m *EmailTemplate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EmailTemplate.Marshal(b, m, deterministic)
}
func (m *EmailTemplate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmailTemplate.Merge(m, src)
}
func (m *EmailTemplate) XXX_Size() int {
	return xxx_messageInfo_EmailTemplate.Size(m)
}
func (m *EmailTemplate) XXX_DiscardUnknown() {
	xxx_messageInfo_EmailTemplate.DiscardUnknown(m)
}

var xxx_messageInfo_EmailTemplate proto.InternalMessageInfo

func (m *EmailTemplate) GetTemplateName() string {
	if m != nil {
		return m.TemplateName
	}
	return ""
}

func (m *EmailTemplate) GetVars() map[string]string {
	if m != nil {
		return m.Vars
	}
	return nil
}

func (m *EmailTemplate) GetEmail() *Email {
	if m != nil {
		return m.Email
	}
	return nil
}

// Email message
type Email struct {
	From                 string   `protobuf:"bytes,1,opt,name=From,proto3" json:"From,omitempty"`
	To                   string   `protobuf:"bytes,2,opt,name=To,proto3" json:"To,omitempty"`
	Subject              string   `protobuf:"bytes,3,opt,name=Subject,proto3" json:"Subject,omitempty"`
	Body                 string   `protobuf:"bytes,4,opt,name=Body,proto3" json:"Body,omitempty"`
	Attachment           []string `protobuf:"bytes,5,rep,name=Attachment,proto3" json:"Attachment,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Email) Reset()         { *m = Email{} }
func (m *Email) String() string { return proto.CompactTextString(m) }
func (*Email) ProtoMessage()    {}
func (*Email) Descriptor() ([]byte, []int) {
	return fileDescriptor_62e234be764a7e86, []int{2}
}

func (m *Email) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Email.Unmarshal(m, b)
}
func (m *Email) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Email.Marshal(b, m, deterministic)
}
func (m *Email) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Email.Merge(m, src)
}
func (m *Email) XXX_Size() int {
	return xxx_messageInfo_Email.Size(m)
}
func (m *Email) XXX_DiscardUnknown() {
	xxx_messageInfo_Email.DiscardUnknown(m)
}

var xxx_messageInfo_Email proto.InternalMessageInfo

func (m *Email) GetFrom() string {
	if m != nil {
		return m.From
	}
	return ""
}

func (m *Email) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}

func (m *Email) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *Email) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

func (m *Email) GetAttachment() []string {
	if m != nil {
		return m.Attachment
	}
	return nil
}

// SMS message
type Sms struct {
	To                   string   `protobuf:"bytes,1,opt,name=To,proto3" json:"To,omitempty"`
	Body                 string   `protobuf:"bytes,2,opt,name=Body,proto3" json:"Body,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Sms) Reset()         { *m = Sms{} }
func (m *Sms) String() string { return proto.CompactTextString(m) }
func (*Sms) ProtoMessage()    {}
func (*Sms) Descriptor() ([]byte, []int) {
	return fileDescriptor_62e234be764a7e86, []int{3}
}

func (m *Sms) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Sms.Unmarshal(m, b)
}
func (m *Sms) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Sms.Marshal(b, m, deterministic)
}
func (m *Sms) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Sms.Merge(m, src)
}
func (m *Sms) XXX_Size() int {
	return xxx_messageInfo_Sms.Size(m)
}
func (m *Sms) XXX_DiscardUnknown() {
	xxx_messageInfo_Sms.DiscardUnknown(m)
}

var xxx_messageInfo_Sms proto.InternalMessageInfo

func (m *Sms) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}

func (m *Sms) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

type Result struct {
	// HTTP STATUS
	// 200 success
	// 500 error
	// 422 validation
	Status               int32    `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Result) Reset()         { *m = Result{} }
func (m *Result) String() string { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()    {}
func (*Result) Descriptor() ([]byte, []int) {
	return fileDescriptor_62e234be764a7e86, []int{4}
}

func (m *Result) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Result.Unmarshal(m, b)
}
func (m *Result) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Result.Marshal(b, m, deterministic)
}
func (m *Result) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Result.Merge(m, src)
}
func (m *Result) XXX_Size() int {
	return xxx_messageInfo_Result.Size(m)
}
func (m *Result) XXX_DiscardUnknown() {
	xxx_messageInfo_Result.DiscardUnknown(m)
}

var xxx_messageInfo_Result proto.InternalMessageInfo

func (m *Result) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *Result) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*OTP)(nil), "NotificationService.OTP")
	proto.RegisterType((*EmailTemplate)(nil), "NotificationService.EmailTemplate")
	proto.RegisterMapType((map[string]string)(nil), "NotificationService.EmailTemplate.VarsEntry")
	proto.RegisterType((*Email)(nil), "NotificationService.Email")
	proto.RegisterType((*Sms)(nil), "NotificationService.Sms")
	proto.RegisterType((*Result)(nil), "NotificationService.Result")
}

func init() { proto.RegisterFile("notification-service.proto", fileDescriptor_62e234be764a7e86) }

var fileDescriptor_62e234be764a7e86 = []byte{
	// 404 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x93, 0xdd, 0x6e, 0x94, 0x40,
	0x14, 0xc7, 0x03, 0x2c, 0x5b, 0x39, 0x55, 0xa3, 0xa3, 0x31, 0x13, 0x4c, 0xcc, 0x86, 0xab, 0x35,
	0x51, 0x62, 0xd6, 0x18, 0x4d, 0xaf, 0xea, 0x47, 0xbd, 0x6c, 0x1b, 0x20, 0xde, 0x4f, 0xd9, 0xa3,
	0xa2, 0x0c, 0xd3, 0xcc, 0x0c, 0x24, 0xbc, 0x99, 0x8f, 0xe3, 0xa3, 0x98, 0x19, 0x06, 0xda, 0x4d,
	0x90, 0xde, 0x9d, 0xaf, 0xff, 0x6f, 0xce, 0x9f, 0x1c, 0x20, 0x6e, 0x84, 0xae, 0xbe, 0x57, 0x25,
	0xd3, 0x95, 0x68, 0x5e, 0x2b, 0x94, 0x5d, 0x55, 0x62, 0x7a, 0x2d, 0x85, 0x16, 0xe4, 0xc9, 0xf9,
	0xad, 0x5e, 0x3e, 0xb4, 0x92, 0x77, 0x10, 0x5c, 0x14, 0x97, 0x24, 0x86, 0x7b, 0x12, 0x4b, 0xac,
	0x3a, 0x94, 0xd4, 0xdb, 0x78, 0xdb, 0x28, 0x9b, 0x72, 0x42, 0x60, 0x55, 0x8a, 0x3d, 0x52, 0xdf,
	0xd6, 0x6d, 0x9c, 0xfc, 0xf5, 0xe0, 0xc1, 0x19, 0x67, 0x55, 0x5d, 0x20, 0xbf, 0xae, 0x99, 0x46,
	0x92, 0xc0, 0x7d, 0xed, 0xe2, 0x73, 0xc6, 0xd1, 0x51, 0x0e, 0x6a, 0xe4, 0x14, 0x56, 0x1d, 0x93,
	0x8a, 0xfa, 0x9b, 0x60, 0x7b, 0xbc, 0x7b, 0x95, 0xce, 0x2c, 0x94, 0x1e, 0x50, 0xd3, 0x6f, 0x4c,
	0xaa, 0xb3, 0x46, 0xcb, 0x3e, 0xb3, 0x4a, 0xf2, 0x06, 0x42, 0x34, 0x03, 0x34, 0xd8, 0x78, 0xdb,
	0xe3, 0x5d, 0xfc, 0x7f, 0x44, 0x36, 0x0c, 0xc6, 0xef, 0x21, 0x9a, 0x20, 0xe4, 0x11, 0x04, 0xbf,
	0xb1, 0x77, 0xbb, 0x99, 0x90, 0x3c, 0x85, 0xb0, 0x63, 0x75, 0x3b, 0xba, 0x1b, 0x92, 0x13, 0xff,
	0x83, 0x97, 0xf4, 0x10, 0x5a, 0x90, 0xf1, 0xff, 0x55, 0x0a, 0xee, 0x54, 0x36, 0x26, 0x0f, 0xc1,
	0x2f, 0x84, 0xd3, 0xf8, 0x85, 0x20, 0x14, 0x8e, 0xf2, 0xf6, 0xea, 0x17, 0x96, 0xda, 0x6e, 0x16,
	0x65, 0x63, 0x6a, 0xd4, 0x9f, 0xc4, 0xbe, 0xa7, 0xab, 0x41, 0x6d, 0x62, 0xf2, 0x02, 0xe0, 0xa3,
	0xd6, 0xac, 0xfc, 0xc9, 0xb1, 0xd1, 0x34, 0xdc, 0x04, 0xdb, 0x28, 0xbb, 0x55, 0x49, 0x5e, 0x42,
	0x90, 0x73, 0xe5, 0x1e, 0xf1, 0xa6, 0x47, 0x46, 0x94, 0x7f, 0x83, 0x4a, 0x4e, 0x60, 0x9d, 0xa1,
	0x6a, 0x6b, 0x4d, 0x9e, 0xc1, 0x5a, 0x69, 0xa6, 0x5b, 0x65, 0x15, 0x61, 0xe6, 0x32, 0xb3, 0x1a,
	0x47, 0xa5, 0xd8, 0x8f, 0xd1, 0xe3, 0x98, 0xee, 0xfe, 0xf8, 0x30, 0x77, 0x13, 0xe4, 0x0b, 0x44,
	0x0a, 0x9b, 0xfd, 0xe0, 0x7e, 0xe1, 0x13, 0xc7, 0xcf, 0x67, 0x7b, 0x6e, 0x9f, 0x53, 0x38, 0x32,
	0x14, 0x63, 0x84, 0xce, 0xce, 0xe5, 0x5c, 0x2d, 0x13, 0x3e, 0x03, 0x38, 0x82, 0x39, 0xd1, 0x79,
	0xc8, 0x45, 0x71, 0xb9, 0x0c, 0x29, 0xe0, 0xf1, 0x64, 0xe6, 0xe6, 0x58, 0xef, 0x3e, 0xbd, 0x45,
	0xea, 0xd5, 0xda, 0xfe, 0x52, 0x6f, 0xff, 0x05, 0x00, 0x00, 0xff, 0xff, 0x1b, 0x42, 0xe5, 0x12,
	0x70, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NotificationServiceClient is the client API for NotificationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NotificationServiceClient interface {
	SendEmail(ctx context.Context, in *Email, opts ...grpc.CallOption) (*Result, error)
	SendSms(ctx context.Context, in *Sms, opts ...grpc.CallOption) (*Result, error)
	SendSmsOTP(ctx context.Context, in *OTP, opts ...grpc.CallOption) (*Result, error)
	SendEmailTemplate(ctx context.Context, in *EmailTemplate, opts ...grpc.CallOption) (*Result, error)
}

type notificationServiceClient struct {
	cc *grpc.ClientConn
}

func NewNotificationServiceClient(cc *grpc.ClientConn) NotificationServiceClient {
	return &notificationServiceClient{cc}
}

func (c *notificationServiceClient) SendEmail(ctx context.Context, in *Email, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/NotificationService.NotificationService/sendEmail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) SendSms(ctx context.Context, in *Sms, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/NotificationService.NotificationService/sendSms", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) SendSmsOTP(ctx context.Context, in *OTP, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/NotificationService.NotificationService/sendSmsOTP", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) SendEmailTemplate(ctx context.Context, in *EmailTemplate, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/NotificationService.NotificationService/sendEmailTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotificationServiceServer is the server API for NotificationService service.
type NotificationServiceServer interface {
	SendEmail(context.Context, *Email) (*Result, error)
	SendSms(context.Context, *Sms) (*Result, error)
	SendSmsOTP(context.Context, *OTP) (*Result, error)
	SendEmailTemplate(context.Context, *EmailTemplate) (*Result, error)
}

// UnimplementedNotificationServiceServer can be embedded to have forward compatible implementations.
type UnimplementedNotificationServiceServer struct {
}

func (*UnimplementedNotificationServiceServer) SendEmail(ctx context.Context, req *Email) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendEmail not implemented")
}
func (*UnimplementedNotificationServiceServer) SendSms(ctx context.Context, req *Sms) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSms not implemented")
}
func (*UnimplementedNotificationServiceServer) SendSmsOTP(ctx context.Context, req *OTP) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSmsOTP not implemented")
}
func (*UnimplementedNotificationServiceServer) SendEmailTemplate(ctx context.Context, req *EmailTemplate) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendEmailTemplate not implemented")
}

func RegisterNotificationServiceServer(s *grpc.Server, srv NotificationServiceServer) {
	s.RegisterService(&_NotificationService_serviceDesc, srv)
}

func _NotificationService_SendEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Email)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).SendEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NotificationService.NotificationService/SendEmail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).SendEmail(ctx, req.(*Email))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_SendSms_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Sms)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).SendSms(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NotificationService.NotificationService/SendSms",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).SendSms(ctx, req.(*Sms))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_SendSmsOTP_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OTP)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).SendSmsOTP(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NotificationService.NotificationService/SendSmsOTP",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).SendSmsOTP(ctx, req.(*OTP))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_SendEmailTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmailTemplate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).SendEmailTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NotificationService.NotificationService/SendEmailTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).SendEmailTemplate(ctx, req.(*EmailTemplate))
	}
	return interceptor(ctx, in, info, handler)
}

var _NotificationService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "NotificationService.NotificationService",
	HandlerType: (*NotificationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "sendEmail",
			Handler:    _NotificationService_SendEmail_Handler,
		},
		{
			MethodName: "sendSms",
			Handler:    _NotificationService_SendSms_Handler,
		},
		{
			MethodName: "sendSmsOTP",
			Handler:    _NotificationService_SendSmsOTP_Handler,
		},
		{
			MethodName: "sendEmailTemplate",
			Handler:    _NotificationService_SendEmailTemplate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "notification-service.proto",
}
