// Code generated by protoc-gen-go. DO NOT EDIT.
// source: envoy/extensions/filters/network/mysql_proxy/v3/mysql_proxy.proto

package envoy_extensions_filters_network_mysql_proxy_v3

import (
	fmt "fmt"
	_ "github.com/cncf/udpa/go/udpa/annotations"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
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

type MySQLProxy struct {
	StatPrefix           string   `protobuf:"bytes,1,opt,name=stat_prefix,json=statPrefix,proto3" json:"stat_prefix,omitempty"`
	AccessLog            string   `protobuf:"bytes,2,opt,name=access_log,json=accessLog,proto3" json:"access_log,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MySQLProxy) Reset()         { *m = MySQLProxy{} }
func (m *MySQLProxy) String() string { return proto.CompactTextString(m) }
func (*MySQLProxy) ProtoMessage()    {}
func (*MySQLProxy) Descriptor() ([]byte, []int) {
	return fileDescriptor_8896af45b7c675fb, []int{0}
}

func (m *MySQLProxy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MySQLProxy.Unmarshal(m, b)
}
func (m *MySQLProxy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MySQLProxy.Marshal(b, m, deterministic)
}
func (m *MySQLProxy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MySQLProxy.Merge(m, src)
}
func (m *MySQLProxy) XXX_Size() int {
	return xxx_messageInfo_MySQLProxy.Size(m)
}
func (m *MySQLProxy) XXX_DiscardUnknown() {
	xxx_messageInfo_MySQLProxy.DiscardUnknown(m)
}

var xxx_messageInfo_MySQLProxy proto.InternalMessageInfo

func (m *MySQLProxy) GetStatPrefix() string {
	if m != nil {
		return m.StatPrefix
	}
	return ""
}

func (m *MySQLProxy) GetAccessLog() string {
	if m != nil {
		return m.AccessLog
	}
	return ""
}

func init() {
	proto.RegisterType((*MySQLProxy)(nil), "envoy.extensions.filters.network.mysql_proxy.v3.MySQLProxy")
}

func init() {
	proto.RegisterFile("envoy/extensions/filters/network/mysql_proxy/v3/mysql_proxy.proto", fileDescriptor_8896af45b7c675fb)
}

var fileDescriptor_8896af45b7c675fb = []byte{
	// 294 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x90, 0xb1, 0x4e, 0xc3, 0x30,
	0x14, 0x45, 0x95, 0x08, 0x15, 0xd5, 0x0c, 0xa0, 0x2c, 0x54, 0x95, 0x8a, 0x0a, 0x53, 0x27, 0x5b,
	0x55, 0x36, 0x50, 0x07, 0x32, 0xb7, 0x52, 0x28, 0x1b, 0x4b, 0x64, 0x52, 0x27, 0x58, 0x04, 0xbf,
	0x60, 0xbb, 0x26, 0xd9, 0x18, 0xf9, 0x05, 0xf8, 0x14, 0x76, 0x24, 0x56, 0x7e, 0x87, 0x09, 0xd9,
	0x0e, 0x6a, 0x25, 0x58, 0xba, 0xe5, 0xe5, 0x5e, 0x9f, 0xe7, 0x63, 0x74, 0xc9, 0x84, 0x81, 0x96,
	0xb0, 0x46, 0x33, 0xa1, 0x38, 0x08, 0x45, 0x0a, 0x5e, 0x69, 0x26, 0x15, 0x11, 0x4c, 0x3f, 0x81,
	0xbc, 0x27, 0x0f, 0xad, 0x7a, 0xac, 0xb2, 0x5a, 0x42, 0xd3, 0x12, 0x13, 0x6f, 0x8f, 0xb8, 0x96,
	0xa0, 0x21, 0x22, 0x0e, 0x81, 0x37, 0x08, 0xdc, 0x21, 0x70, 0x87, 0xc0, 0xdb, 0x67, 0x4c, 0x3c,
	0x1c, 0xad, 0x57, 0x35, 0x25, 0x54, 0x08, 0xd0, 0x54, 0xbb, 0x9d, 0x4a, 0x53, 0xbd, 0x56, 0x9e,
	0x37, 0x3c, 0xfd, 0x13, 0x1b, 0x26, 0x2d, 0x98, 0x8b, 0xb2, 0xab, 0x1c, 0x1b, 0x5a, 0xf1, 0x15,
	0xd5, 0x8c, 0xfc, 0x7e, 0xf8, 0xe0, 0xec, 0x35, 0x40, 0x68, 0xd1, 0x5e, 0x5f, 0xcd, 0x53, 0xbb,
	0x2c, 0x9a, 0xa0, 0x03, 0x8b, 0xce, 0x6a, 0xc9, 0x0a, 0xde, 0x0c, 0x82, 0x71, 0x30, 0xe9, 0x27,
	0xfb, 0xdf, 0xc9, 0x9e, 0x0c, 0xc7, 0xc1, 0x12, 0xd9, 0x2c, 0x75, 0x51, 0x34, 0x42, 0x88, 0xe6,
	0x39, 0x53, 0x2a, 0xab, 0xa0, 0x1c, 0x84, 0xb6, 0xb8, 0xec, 0xfb, 0x3f, 0x73, 0x28, 0xcf, 0x93,
	0xb7, 0x8f, 0x97, 0x93, 0x19, 0xba, 0xf0, 0xaa, 0x39, 0x88, 0x82, 0x97, 0x9d, 0xe6, 0xff, 0x96,
	0x53, 0x5a, 0xd5, 0x77, 0x74, 0x8a, 0x37, 0x97, 0x49, 0x6e, 0xde, 0x9f, 0x3f, 0xbf, 0x7a, 0xe1,
	0x51, 0x88, 0x66, 0x1c, 0xb0, 0x23, 0xf9, 0xf2, 0x8e, 0xef, 0x97, 0x1c, 0x2e, 0xec, 0xec, 0xa0,
	0xa9, 0xb5, 0x4e, 0x83, 0xdb, 0x9e, 0xd3, 0x8f, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0xfa, 0xe7,
	0x24, 0xba, 0xcf, 0x01, 0x00, 0x00,
}
