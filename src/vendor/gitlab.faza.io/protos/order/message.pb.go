// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message.proto

package ordersrv

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type ResponseMetadata struct {
	Total                uint32   `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Page                 uint32   `protobuf:"varint,2,opt,name=page,proto3" json:"page,omitempty"`
	PerPage              uint32   `protobuf:"varint,3,opt,name=perPage,proto3" json:"perPage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResponseMetadata) Reset()         { *m = ResponseMetadata{} }
func (m *ResponseMetadata) String() string { return proto.CompactTextString(m) }
func (*ResponseMetadata) ProtoMessage()    {}
func (*ResponseMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{0}
}

func (m *ResponseMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResponseMetadata.Unmarshal(m, b)
}
func (m *ResponseMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResponseMetadata.Marshal(b, m, deterministic)
}
func (m *ResponseMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResponseMetadata.Merge(m, src)
}
func (m *ResponseMetadata) XXX_Size() int {
	return xxx_messageInfo_ResponseMetadata.Size(m)
}
func (m *ResponseMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_ResponseMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_ResponseMetadata proto.InternalMessageInfo

func (m *ResponseMetadata) GetTotal() uint32 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *ResponseMetadata) GetPage() uint32 {
	if m != nil {
		return m.Page
	}
	return 0
}

func (m *ResponseMetadata) GetPerPage() uint32 {
	if m != nil {
		return m.PerPage
	}
	return 0
}

type MessageResponse struct {
	Entity               string            `protobuf:"bytes,1,opt,name=entity,proto3" json:"entity,omitempty"`
	Id                   string            `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Meta                 *ResponseMetadata `protobuf:"bytes,3,opt,name=meta,proto3" json:"meta,omitempty"`
	Data                 *any.Any          `protobuf:"bytes,4,opt,name=Data,proto3" json:"Data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *MessageResponse) Reset()         { *m = MessageResponse{} }
func (m *MessageResponse) String() string { return proto.CompactTextString(m) }
func (*MessageResponse) ProtoMessage()    {}
func (*MessageResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{1}
}

func (m *MessageResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MessageResponse.Unmarshal(m, b)
}
func (m *MessageResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MessageResponse.Marshal(b, m, deterministic)
}
func (m *MessageResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageResponse.Merge(m, src)
}
func (m *MessageResponse) XXX_Size() int {
	return xxx_messageInfo_MessageResponse.Size(m)
}
func (m *MessageResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MessageResponse proto.InternalMessageInfo

func (m *MessageResponse) GetEntity() string {
	if m != nil {
		return m.Entity
	}
	return ""
}

func (m *MessageResponse) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *MessageResponse) GetMeta() *ResponseMetadata {
	if m != nil {
		return m.Meta
	}
	return nil
}

func (m *MessageResponse) GetData() *any.Any {
	if m != nil {
		return m.Data
	}
	return nil
}

type MetaFilter struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Opt                  string   `protobuf:"bytes,2,opt,name=opt,proto3" json:"opt,omitempty"`
	Value                string   `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MetaFilter) Reset()         { *m = MetaFilter{} }
func (m *MetaFilter) String() string { return proto.CompactTextString(m) }
func (*MetaFilter) ProtoMessage()    {}
func (*MetaFilter) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{2}
}

func (m *MetaFilter) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MetaFilter.Unmarshal(m, b)
}
func (m *MetaFilter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MetaFilter.Marshal(b, m, deterministic)
}
func (m *MetaFilter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MetaFilter.Merge(m, src)
}
func (m *MetaFilter) XXX_Size() int {
	return xxx_messageInfo_MetaFilter.Size(m)
}
func (m *MetaFilter) XXX_DiscardUnknown() {
	xxx_messageInfo_MetaFilter.DiscardUnknown(m)
}

var xxx_messageInfo_MetaFilter proto.InternalMessageInfo

func (m *MetaFilter) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *MetaFilter) GetOpt() string {
	if m != nil {
		return m.Opt
	}
	return ""
}

func (m *MetaFilter) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type MetaSorts struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Direction            int32    `protobuf:"varint,2,opt,name=direction,proto3" json:"direction,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MetaSorts) Reset()         { *m = MetaSorts{} }
func (m *MetaSorts) String() string { return proto.CompactTextString(m) }
func (*MetaSorts) ProtoMessage()    {}
func (*MetaSorts) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{3}
}

func (m *MetaSorts) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MetaSorts.Unmarshal(m, b)
}
func (m *MetaSorts) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MetaSorts.Marshal(b, m, deterministic)
}
func (m *MetaSorts) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MetaSorts.Merge(m, src)
}
func (m *MetaSorts) XXX_Size() int {
	return xxx_messageInfo_MetaSorts.Size(m)
}
func (m *MetaSorts) XXX_DiscardUnknown() {
	xxx_messageInfo_MetaSorts.DiscardUnknown(m)
}

var xxx_messageInfo_MetaSorts proto.InternalMessageInfo

func (m *MetaSorts) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *MetaSorts) GetDirection() int32 {
	if m != nil {
		return m.Direction
	}
	return 0
}

type RequestMetadata struct {
	Page                 uint32        `protobuf:"varint,1,opt,name=page,proto3" json:"page,omitempty"`
	PerPage              uint32        `protobuf:"varint,2,opt,name=perPage,proto3" json:"perPage,omitempty"`
	Sorts                []*MetaSorts  `protobuf:"bytes,3,rep,name=sorts,proto3" json:"sorts,omitempty"`
	Filters              []*MetaFilter `protobuf:"bytes,4,rep,name=filters,proto3" json:"filters,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *RequestMetadata) Reset()         { *m = RequestMetadata{} }
func (m *RequestMetadata) String() string { return proto.CompactTextString(m) }
func (*RequestMetadata) ProtoMessage()    {}
func (*RequestMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{4}
}

func (m *RequestMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RequestMetadata.Unmarshal(m, b)
}
func (m *RequestMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RequestMetadata.Marshal(b, m, deterministic)
}
func (m *RequestMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RequestMetadata.Merge(m, src)
}
func (m *RequestMetadata) XXX_Size() int {
	return xxx_messageInfo_RequestMetadata.Size(m)
}
func (m *RequestMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_RequestMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_RequestMetadata proto.InternalMessageInfo

func (m *RequestMetadata) GetPage() uint32 {
	if m != nil {
		return m.Page
	}
	return 0
}

func (m *RequestMetadata) GetPerPage() uint32 {
	if m != nil {
		return m.PerPage
	}
	return 0
}

func (m *RequestMetadata) GetSorts() []*MetaSorts {
	if m != nil {
		return m.Sorts
	}
	return nil
}

func (m *RequestMetadata) GetFilters() []*MetaFilter {
	if m != nil {
		return m.Filters
	}
	return nil
}

type MessageRequest struct {
	OrderId              string               `protobuf:"bytes,1,opt,name=orderId,proto3" json:"orderId,omitempty"`
	Time                 *timestamp.Timestamp `protobuf:"bytes,2,opt,name=time,proto3" json:"time,omitempty"`
	Meta                 *RequestMetadata     `protobuf:"bytes,3,opt,name=meta,proto3" json:"meta,omitempty"`
	Data                 *any.Any             `protobuf:"bytes,4,opt,name=Data,proto3" json:"Data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *MessageRequest) Reset()         { *m = MessageRequest{} }
func (m *MessageRequest) String() string { return proto.CompactTextString(m) }
func (*MessageRequest) ProtoMessage()    {}
func (*MessageRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{5}
}

func (m *MessageRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MessageRequest.Unmarshal(m, b)
}
func (m *MessageRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MessageRequest.Marshal(b, m, deterministic)
}
func (m *MessageRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageRequest.Merge(m, src)
}
func (m *MessageRequest) XXX_Size() int {
	return xxx_messageInfo_MessageRequest.Size(m)
}
func (m *MessageRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MessageRequest proto.InternalMessageInfo

func (m *MessageRequest) GetOrderId() string {
	if m != nil {
		return m.OrderId
	}
	return ""
}

func (m *MessageRequest) GetTime() *timestamp.Timestamp {
	if m != nil {
		return m.Time
	}
	return nil
}

func (m *MessageRequest) GetMeta() *RequestMetadata {
	if m != nil {
		return m.Meta
	}
	return nil
}

func (m *MessageRequest) GetData() *any.Any {
	if m != nil {
		return m.Data
	}
	return nil
}

type ErrorDetails struct {
	Validation           []*ValidationErr `protobuf:"bytes,1,rep,name=validation,proto3" json:"validation,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *ErrorDetails) Reset()         { *m = ErrorDetails{} }
func (m *ErrorDetails) String() string { return proto.CompactTextString(m) }
func (*ErrorDetails) ProtoMessage()    {}
func (*ErrorDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{6}
}

func (m *ErrorDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ErrorDetails.Unmarshal(m, b)
}
func (m *ErrorDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ErrorDetails.Marshal(b, m, deterministic)
}
func (m *ErrorDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ErrorDetails.Merge(m, src)
}
func (m *ErrorDetails) XXX_Size() int {
	return xxx_messageInfo_ErrorDetails.Size(m)
}
func (m *ErrorDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_ErrorDetails.DiscardUnknown(m)
}

var xxx_messageInfo_ErrorDetails proto.InternalMessageInfo

func (m *ErrorDetails) GetValidation() []*ValidationErr {
	if m != nil {
		return m.Validation
	}
	return nil
}

type ValidationErr struct {
	Field                string   `protobuf:"bytes,1,opt,name=field,proto3" json:"field,omitempty"`
	Desc                 string   `protobuf:"bytes,2,opt,name=desc,proto3" json:"desc,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidationErr) Reset()         { *m = ValidationErr{} }
func (m *ValidationErr) String() string { return proto.CompactTextString(m) }
func (*ValidationErr) ProtoMessage()    {}
func (*ValidationErr) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{7}
}

func (m *ValidationErr) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidationErr.Unmarshal(m, b)
}
func (m *ValidationErr) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidationErr.Marshal(b, m, deterministic)
}
func (m *ValidationErr) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidationErr.Merge(m, src)
}
func (m *ValidationErr) XXX_Size() int {
	return xxx_messageInfo_ValidationErr.Size(m)
}
func (m *ValidationErr) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidationErr.DiscardUnknown(m)
}

var xxx_messageInfo_ValidationErr proto.InternalMessageInfo

func (m *ValidationErr) GetField() string {
	if m != nil {
		return m.Field
	}
	return ""
}

func (m *ValidationErr) GetDesc() string {
	if m != nil {
		return m.Desc
	}
	return ""
}

func init() {
	proto.RegisterType((*ResponseMetadata)(nil), "ordersrv.ResponseMetadata")
	proto.RegisterType((*MessageResponse)(nil), "ordersrv.MessageResponse")
	proto.RegisterType((*MetaFilter)(nil), "ordersrv.MetaFilter")
	proto.RegisterType((*MetaSorts)(nil), "ordersrv.MetaSorts")
	proto.RegisterType((*RequestMetadata)(nil), "ordersrv.RequestMetadata")
	proto.RegisterType((*MessageRequest)(nil), "ordersrv.MessageRequest")
	proto.RegisterType((*ErrorDetails)(nil), "ordersrv.ErrorDetails")
	proto.RegisterType((*ValidationErr)(nil), "ordersrv.ValidationErr")
}

func init() { proto.RegisterFile("message.proto", fileDescriptor_33c57e4bae7b9afd) }

var fileDescriptor_33c57e4bae7b9afd = []byte{
	// 463 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x96, 0xf3, 0xd3, 0xe2, 0x29, 0x69, 0xab, 0x25, 0x02, 0x37, 0x42, 0xa2, 0xf2, 0x29, 0x1c,
	0x70, 0xa5, 0x72, 0x40, 0x1c, 0x38, 0x20, 0xb5, 0xfc, 0x1c, 0x2a, 0xa1, 0x05, 0xf5, 0xbe, 0xad,
	0x27, 0xd1, 0x4a, 0xb6, 0xd7, 0xec, 0x4e, 0x22, 0xe5, 0x39, 0xb8, 0xf1, 0x1a, 0xbc, 0x20, 0xda,
	0x59, 0x6f, 0x52, 0x47, 0xbd, 0xf4, 0x36, 0x33, 0xfb, 0x79, 0xfc, 0xfd, 0x0c, 0x4c, 0x6a, 0x74,
	0x4e, 0x2d, 0xb1, 0x68, 0xad, 0x21, 0x23, 0x9e, 0x19, 0x5b, 0xa2, 0x75, 0x76, 0x3d, 0x3b, 0x5b,
	0x1a, 0xb3, 0xac, 0xf0, 0x82, 0xe7, 0x77, 0xab, 0xc5, 0x85, 0x6a, 0x36, 0x01, 0x34, 0x7b, 0xb3,
	0xff, 0x44, 0xba, 0x46, 0x47, 0xaa, 0x6e, 0x03, 0x20, 0xbf, 0x85, 0x53, 0x89, 0xae, 0x35, 0x8d,
	0xc3, 0x1b, 0x24, 0x55, 0x2a, 0x52, 0x62, 0x0a, 0x63, 0x32, 0xa4, 0xaa, 0x2c, 0x39, 0x4f, 0xe6,
	0x13, 0x19, 0x1a, 0x21, 0x60, 0xd4, 0xaa, 0x25, 0x66, 0x03, 0x1e, 0x72, 0x2d, 0x32, 0x38, 0x6c,
	0xd1, 0xfe, 0xf0, 0xe3, 0x21, 0x8f, 0x63, 0x9b, 0xff, 0x49, 0xe0, 0xe4, 0x26, 0xf0, 0x8d, 0xfb,
	0xc5, 0x4b, 0x38, 0xc0, 0x86, 0x34, 0x6d, 0x78, 0x71, 0x2a, 0xbb, 0x4e, 0x1c, 0xc3, 0x40, 0x97,
	0xbc, 0x37, 0x95, 0x03, 0x5d, 0x8a, 0x02, 0x46, 0x35, 0x92, 0xe2, 0x95, 0x47, 0x97, 0xb3, 0x22,
	0x0a, 0x2d, 0xf6, 0x99, 0x4a, 0xc6, 0x89, 0x39, 0x8c, 0xae, 0x14, 0xa9, 0x6c, 0xc4, 0xf8, 0x69,
	0x11, 0x34, 0x17, 0x51, 0x73, 0xf1, 0xb9, 0xd9, 0x48, 0x46, 0xe4, 0xdf, 0x00, 0xfc, 0xb7, 0x5f,
	0x74, 0x45, 0x68, 0xbd, 0xa2, 0x46, 0xd5, 0xd8, 0xb1, 0xe1, 0x5a, 0x9c, 0xc2, 0xd0, 0xb4, 0xd4,
	0x91, 0xf1, 0xa5, 0x77, 0x63, 0xad, 0xaa, 0x55, 0x50, 0x98, 0xca, 0xd0, 0xe4, 0x9f, 0x20, 0xf5,
	0x9b, 0x7e, 0x1a, 0x4b, 0xee, 0xd1, 0x45, 0xaf, 0x21, 0x2d, 0xb5, 0xc5, 0x7b, 0xd2, 0xa6, 0xe1,
	0x75, 0x63, 0xb9, 0x1b, 0xe4, 0x7f, 0x13, 0x38, 0x91, 0xf8, 0x7b, 0x85, 0x8e, 0xb6, 0xb6, 0x47,
	0x83, 0x93, 0xc7, 0x0d, 0x1e, 0xf4, 0x0c, 0x16, 0x6f, 0x61, 0xec, 0xfc, 0xcf, 0xb3, 0xe1, 0xf9,
	0x70, 0x7e, 0x74, 0xf9, 0x62, 0xe7, 0xd2, 0x96, 0x97, 0x0c, 0x08, 0x51, 0xc0, 0xe1, 0x82, 0x15,
	0xbb, 0x6c, 0xc4, 0xe0, 0x69, 0x1f, 0x1c, 0xec, 0x90, 0x11, 0x94, 0xff, 0x4b, 0xe0, 0x78, 0x9b,
	0x1d, 0x73, 0xf4, 0x3c, 0xf8, 0x93, 0xef, 0x65, 0x27, 0x32, 0xb6, 0x3e, 0x2c, 0x7f, 0x53, 0x4c,
	0xcf, 0x87, 0xb5, 0x6f, 0xfe, 0xaf, 0x78, 0x70, 0x92, 0x71, 0xe2, 0x5d, 0x2f, 0xdc, 0xb3, 0x87,
	0xe1, 0xf6, 0xec, 0x78, 0x72, 0xb6, 0x5f, 0xe1, 0xf9, 0xb5, 0xb5, 0xc6, 0x5e, 0x21, 0x29, 0x5d,
	0x39, 0xf1, 0x01, 0x60, 0xad, 0x2a, 0x5d, 0x2a, 0x4e, 0x20, 0x61, 0xe1, 0xaf, 0x76, 0xbf, 0xbb,
	0xdd, 0xbe, 0x5d, 0x5b, 0x2b, 0x1f, 0x40, 0xf3, 0x8f, 0x30, 0xe9, 0x3d, 0xfa, 0x0b, 0x58, 0x68,
	0xac, 0xa2, 0xf4, 0xd0, 0xf8, 0xb8, 0x4a, 0x74, 0xf7, 0xdd, 0xa9, 0x70, 0x7d, 0x77, 0xc0, 0xbc,
	0xde, 0xff, 0x0f, 0x00, 0x00, 0xff, 0xff, 0xd4, 0x84, 0x20, 0x48, 0xab, 0x03, 0x00, 0x00,
}