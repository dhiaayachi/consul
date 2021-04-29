// Code generated by protoc-gen-go. DO NOT EDIT.
// source: envoy/config/accesslog/v2/file.proto

package envoy_config_accesslog_v2

import (
	fmt "fmt"
	_ "github.com/cncf/udpa/go/udpa/annotations"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
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

type FileAccessLog struct {
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Types that are valid to be assigned to AccessLogFormat:
	//	*FileAccessLog_Format
	//	*FileAccessLog_JsonFormat
	//	*FileAccessLog_TypedJsonFormat
	AccessLogFormat      isFileAccessLog_AccessLogFormat `protobuf_oneof:"access_log_format"`
	XXX_NoUnkeyedLiteral struct{}                        `json:"-"`
	XXX_unrecognized     []byte                          `json:"-"`
	XXX_sizecache        int32                           `json:"-"`
}

func (m *FileAccessLog) Reset()         { *m = FileAccessLog{} }
func (m *FileAccessLog) String() string { return proto.CompactTextString(m) }
func (*FileAccessLog) ProtoMessage()    {}
func (*FileAccessLog) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb42a04cfa71ce3c, []int{0}
}

func (m *FileAccessLog) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileAccessLog.Unmarshal(m, b)
}
func (m *FileAccessLog) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileAccessLog.Marshal(b, m, deterministic)
}
func (m *FileAccessLog) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileAccessLog.Merge(m, src)
}
func (m *FileAccessLog) XXX_Size() int {
	return xxx_messageInfo_FileAccessLog.Size(m)
}
func (m *FileAccessLog) XXX_DiscardUnknown() {
	xxx_messageInfo_FileAccessLog.DiscardUnknown(m)
}

var xxx_messageInfo_FileAccessLog proto.InternalMessageInfo

func (m *FileAccessLog) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

type isFileAccessLog_AccessLogFormat interface {
	isFileAccessLog_AccessLogFormat()
}

type FileAccessLog_Format struct {
	Format string `protobuf:"bytes,2,opt,name=format,proto3,oneof"`
}

type FileAccessLog_JsonFormat struct {
	JsonFormat *_struct.Struct `protobuf:"bytes,3,opt,name=json_format,json=jsonFormat,proto3,oneof"`
}

type FileAccessLog_TypedJsonFormat struct {
	TypedJsonFormat *_struct.Struct `protobuf:"bytes,4,opt,name=typed_json_format,json=typedJsonFormat,proto3,oneof"`
}

func (*FileAccessLog_Format) isFileAccessLog_AccessLogFormat() {}

func (*FileAccessLog_JsonFormat) isFileAccessLog_AccessLogFormat() {}

func (*FileAccessLog_TypedJsonFormat) isFileAccessLog_AccessLogFormat() {}

func (m *FileAccessLog) GetAccessLogFormat() isFileAccessLog_AccessLogFormat {
	if m != nil {
		return m.AccessLogFormat
	}
	return nil
}

func (m *FileAccessLog) GetFormat() string {
	if x, ok := m.GetAccessLogFormat().(*FileAccessLog_Format); ok {
		return x.Format
	}
	return ""
}

func (m *FileAccessLog) GetJsonFormat() *_struct.Struct {
	if x, ok := m.GetAccessLogFormat().(*FileAccessLog_JsonFormat); ok {
		return x.JsonFormat
	}
	return nil
}

func (m *FileAccessLog) GetTypedJsonFormat() *_struct.Struct {
	if x, ok := m.GetAccessLogFormat().(*FileAccessLog_TypedJsonFormat); ok {
		return x.TypedJsonFormat
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*FileAccessLog) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*FileAccessLog_Format)(nil),
		(*FileAccessLog_JsonFormat)(nil),
		(*FileAccessLog_TypedJsonFormat)(nil),
	}
}

func init() {
	proto.RegisterType((*FileAccessLog)(nil), "envoy.config.accesslog.v2.FileAccessLog")
}

func init() {
	proto.RegisterFile("envoy/config/accesslog/v2/file.proto", fileDescriptor_bb42a04cfa71ce3c)
}

var fileDescriptor_bb42a04cfa71ce3c = []byte{
	// 349 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0xcf, 0x4e, 0x32, 0x31,
	0x14, 0xc5, 0x19, 0x3e, 0x3e, 0x0c, 0x25, 0x46, 0x19, 0x17, 0x20, 0xfe, 0x09, 0x31, 0x26, 0xe0,
	0xa6, 0x4d, 0x60, 0xe7, 0xce, 0x49, 0x24, 0xc4, 0xb8, 0x20, 0xf8, 0x00, 0xa4, 0x0c, 0x9d, 0x5a,
	0x33, 0xf4, 0x4e, 0xda, 0xce, 0x04, 0x76, 0xbe, 0x81, 0x5b, 0x9f, 0xc5, 0x27, 0x70, 0xeb, 0x0b,
	0xf8, 0x10, 0x2e, 0x5d, 0x18, 0x33, 0xed, 0x80, 0x26, 0xc6, 0xb8, 0x9b, 0xe9, 0xf9, 0x9d, 0x93,
	0x7b, 0xcf, 0x45, 0xa7, 0x4c, 0x66, 0xb0, 0x22, 0x21, 0xc8, 0x48, 0x70, 0x42, 0xc3, 0x90, 0x69,
	0x1d, 0x03, 0x27, 0x59, 0x9f, 0x44, 0x22, 0x66, 0x38, 0x51, 0x60, 0xc0, 0xdf, 0xb7, 0x14, 0x76,
	0x14, 0xde, 0x50, 0x38, 0xeb, 0xb7, 0x0f, 0x39, 0x00, 0x8f, 0x19, 0xb1, 0xe0, 0x2c, 0x8d, 0x88,
	0x36, 0x2a, 0x0d, 0x8d, 0x33, 0xb6, 0x8f, 0xd3, 0x79, 0x42, 0x09, 0x95, 0x12, 0x0c, 0x35, 0x02,
	0xa4, 0x26, 0x0b, 0xc1, 0x15, 0x35, 0x45, 0x70, 0xfb, 0xe8, 0x87, 0xae, 0x0d, 0x35, 0xa9, 0x2e,
	0xe4, 0x66, 0x46, 0x63, 0x31, 0xa7, 0x86, 0x91, 0xf5, 0x87, 0x13, 0x4e, 0x5e, 0x3d, 0xb4, 0x3d,
	0x14, 0x31, 0xbb, 0xb0, 0xa3, 0x5c, 0x03, 0xf7, 0x0f, 0x50, 0x25, 0xa1, 0xe6, 0xb6, 0xe5, 0x75,
	0xbc, 0x5e, 0x2d, 0xd8, 0x7a, 0x0f, 0x2a, 0xaa, 0xdc, 0xf1, 0x26, 0xf6, 0xd1, 0x6f, 0xa1, 0x6a,
	0x04, 0x6a, 0x41, 0x4d, 0xab, 0x9c, 0xcb, 0xa3, 0xd2, 0xa4, 0xf8, 0xf7, 0xcf, 0x51, 0xfd, 0x4e,
	0x83, 0x9c, 0x16, 0xf2, 0xbf, 0x8e, 0xd7, 0xab, 0xf7, 0x9b, 0xd8, 0x2d, 0x85, 0xd7, 0x4b, 0xe1,
	0x1b, 0xbb, 0xd4, 0xa8, 0x34, 0x41, 0x39, 0x3d, 0x74, 0xde, 0x4b, 0xd4, 0x30, 0xab, 0x84, 0xcd,
	0xa7, 0xdf, 0x13, 0x2a, 0x7f, 0x25, 0xec, 0x58, 0xcf, 0xd5, 0x26, 0x26, 0xd8, 0x43, 0x0d, 0xd7,
	0xe8, 0x34, 0x06, 0x5e, 0xc4, 0x04, 0x8b, 0xb7, 0xc7, 0x8f, 0x87, 0xff, 0x67, 0x7e, 0xd7, 0x35,
	0xcf, 0x96, 0x86, 0x49, 0x9d, 0x17, 0x84, 0xbf, 0x58, 0xce, 0x94, 0xc6, 0xf6, 0x4a, 0xd9, 0xe0,
	0xe9, 0xfe, 0xf9, 0xa5, 0x5a, 0xde, 0xf5, 0x50, 0x57, 0x00, 0xb6, 0x9e, 0x44, 0xc1, 0x72, 0x85,
	0x7f, 0x3d, 0x5c, 0x50, 0xcb, 0xfb, 0x1b, 0xe7, 0xe3, 0x8d, 0xbd, 0x59, 0xd5, 0xce, 0x39, 0xf8,
	0x0c, 0x00, 0x00, 0xff, 0xff, 0x11, 0x00, 0x44, 0x6a, 0x0f, 0x02, 0x00, 0x00,
}
