// Code generated by protoc-gen-go. DO NOT EDIT.
// source: kvetch/api/v1/key_value.proto

package apiv1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

// KeyValue is a key value object.
type KeyValue struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeyValue) Reset()         { *m = KeyValue{} }
func (m *KeyValue) String() string { return proto.CompactTextString(m) }
func (*KeyValue) ProtoMessage()    {}
func (*KeyValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_da126830bd373ffc, []int{0}
}

func (m *KeyValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeyValue.Unmarshal(m, b)
}
func (m *KeyValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeyValue.Marshal(b, m, deterministic)
}
func (m *KeyValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeyValue.Merge(m, src)
}
func (m *KeyValue) XXX_Size() int {
	return xxx_messageInfo_KeyValue.Size(m)
}
func (m *KeyValue) XXX_DiscardUnknown() {
	xxx_messageInfo_KeyValue.DiscardUnknown(m)
}

var xxx_messageInfo_KeyValue proto.InternalMessageInfo

func (m *KeyValue) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KeyValue) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto.RegisterType((*KeyValue)(nil), "kvetch.api.v1.KeyValue")
}

func init() {
	proto.RegisterFile("kvetch/api/v1/key_value.proto", fileDescriptor_da126830bd373ffc)
}

var fileDescriptor_da126830bd373ffc = []byte{
	// 156 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0xcd, 0x2e, 0x4b, 0x2d,
	0x49, 0xce, 0xd0, 0x4f, 0x2c, 0xc8, 0xd4, 0x2f, 0x33, 0xd4, 0xcf, 0x4e, 0xad, 0x8c, 0x2f, 0x4b,
	0xcc, 0x29, 0x4d, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x85, 0x48, 0xeb, 0x25, 0x16,
	0x64, 0xea, 0x95, 0x19, 0x2a, 0x19, 0x71, 0x71, 0x78, 0xa7, 0x56, 0x86, 0x81, 0x14, 0x08, 0x09,
	0x70, 0x31, 0x67, 0xa7, 0x56, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06, 0x81, 0x98, 0x42, 0x22,
	0x5c, 0xac, 0x60, 0xbd, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0x3c, 0x41, 0x10, 0x8e, 0x93, 0x23, 0x97,
	0x60, 0x72, 0x7e, 0xae, 0x1e, 0x8a, 0x41, 0x4e, 0xbc, 0x30, 0x63, 0x02, 0x40, 0xd6, 0x04, 0x30,
	0x46, 0xb1, 0x26, 0x16, 0x64, 0x96, 0x19, 0x2e, 0x62, 0x62, 0xf6, 0x76, 0x8c, 0x58, 0xc5, 0xc4,
	0xeb, 0x0d, 0x51, 0xed, 0x58, 0x90, 0xa9, 0x17, 0x66, 0x98, 0xc4, 0x06, 0x76, 0x8c, 0x31, 0x20,
	0x00, 0x00, 0xff, 0xff, 0xe6, 0x5b, 0xc7, 0x5d, 0xad, 0x00, 0x00, 0x00,
}
