// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: pbcatalog/v2beta1/failover_policy.proto

package catalogv2beta1

import (
	pbresource "github.com/hashicorp/consul/proto-public/pbresource"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type FailoverMode int32

const (
	FailoverMode_FAILOVER_MODE_UNSPECIFIED       FailoverMode = 0
	FailoverMode_FAILOVER_MODE_SEQUENTIAL        FailoverMode = 1
	FailoverMode_FAILOVER_MODE_ORDER_BY_LOCALITY FailoverMode = 2
)

// Enum value maps for FailoverMode.
var (
	FailoverMode_name = map[int32]string{
		0: "FAILOVER_MODE_UNSPECIFIED",
		1: "FAILOVER_MODE_SEQUENTIAL",
		2: "FAILOVER_MODE_ORDER_BY_LOCALITY",
	}
	FailoverMode_value = map[string]int32{
		"FAILOVER_MODE_UNSPECIFIED":       0,
		"FAILOVER_MODE_SEQUENTIAL":        1,
		"FAILOVER_MODE_ORDER_BY_LOCALITY": 2,
	}
)

func (x FailoverMode) Enum() *FailoverMode {
	p := new(FailoverMode)
	*p = x
	return p
}

func (x FailoverMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FailoverMode) Descriptor() protoreflect.EnumDescriptor {
	return file_pbcatalog_v2beta1_failover_policy_proto_enumTypes[0].Descriptor()
}

func (FailoverMode) Type() protoreflect.EnumType {
	return &file_pbcatalog_v2beta1_failover_policy_proto_enumTypes[0]
}

func (x FailoverMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FailoverMode.Descriptor instead.
func (FailoverMode) EnumDescriptor() ([]byte, []int) {
	return file_pbcatalog_v2beta1_failover_policy_proto_rawDescGZIP(), []int{0}
}

// This is a Resource type.
type FailoverPolicy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Config defines failover for any named port not present in PortConfigs.
	Config *FailoverConfig `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	// PortConfigs defines failover for a specific port on this service and takes
	// precedence over Config.
	PortConfigs map[string]*FailoverConfig `protobuf:"bytes,2,rep,name=port_configs,json=portConfigs,proto3" json:"port_configs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *FailoverPolicy) Reset() {
	*x = FailoverPolicy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FailoverPolicy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FailoverPolicy) ProtoMessage() {}

func (x *FailoverPolicy) ProtoReflect() protoreflect.Message {
	mi := &file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FailoverPolicy.ProtoReflect.Descriptor instead.
func (*FailoverPolicy) Descriptor() ([]byte, []int) {
	return file_pbcatalog_v2beta1_failover_policy_proto_rawDescGZIP(), []int{0}
}

func (x *FailoverPolicy) GetConfig() *FailoverConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *FailoverPolicy) GetPortConfigs() map[string]*FailoverConfig {
	if x != nil {
		return x.PortConfigs
	}
	return nil
}

type FailoverConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Destinations specifies a fixed list of failover destinations to try. We
	// never try a destination multiple times, so those are subtracted from this
	// list before proceeding.
	Destinations []*FailoverDestination `protobuf:"bytes,1,rep,name=destinations,proto3" json:"destinations,omitempty"`
	// Mode specifies the type of failover that will be performed. Valid values are
	// "sequential", "" (equivalent to "sequential") and "order-by-locality".
	Mode    FailoverMode `protobuf:"varint,2,opt,name=mode,proto3,enum=hashicorp.consul.catalog.v2beta1.FailoverMode" json:"mode,omitempty"`
	Regions []string     `protobuf:"bytes,3,rep,name=regions,proto3" json:"regions,omitempty"`
	// SamenessGroup specifies the sameness group to failover to.
	SamenessGroup string `protobuf:"bytes,4,opt,name=sameness_group,json=samenessGroup,proto3" json:"sameness_group,omitempty"`
}

func (x *FailoverConfig) Reset() {
	*x = FailoverConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FailoverConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FailoverConfig) ProtoMessage() {}

func (x *FailoverConfig) ProtoReflect() protoreflect.Message {
	mi := &file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FailoverConfig.ProtoReflect.Descriptor instead.
func (*FailoverConfig) Descriptor() ([]byte, []int) {
	return file_pbcatalog_v2beta1_failover_policy_proto_rawDescGZIP(), []int{1}
}

func (x *FailoverConfig) GetDestinations() []*FailoverDestination {
	if x != nil {
		return x.Destinations
	}
	return nil
}

func (x *FailoverConfig) GetMode() FailoverMode {
	if x != nil {
		return x.Mode
	}
	return FailoverMode_FAILOVER_MODE_UNSPECIFIED
}

func (x *FailoverConfig) GetRegions() []string {
	if x != nil {
		return x.Regions
	}
	return nil
}

func (x *FailoverConfig) GetSamenessGroup() string {
	if x != nil {
		return x.SamenessGroup
	}
	return ""
}

type FailoverDestination struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// This must be a Service.
	Ref *pbresource.Reference `protobuf:"bytes,1,opt,name=ref,proto3" json:"ref,omitempty"`
	// TODO: what should an empty port mean?
	Port       string `protobuf:"bytes,2,opt,name=port,proto3" json:"port,omitempty"`
	Datacenter string `protobuf:"bytes,3,opt,name=datacenter,proto3" json:"datacenter,omitempty"`
}

func (x *FailoverDestination) Reset() {
	*x = FailoverDestination{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FailoverDestination) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FailoverDestination) ProtoMessage() {}

func (x *FailoverDestination) ProtoReflect() protoreflect.Message {
	mi := &file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FailoverDestination.ProtoReflect.Descriptor instead.
func (*FailoverDestination) Descriptor() ([]byte, []int) {
	return file_pbcatalog_v2beta1_failover_policy_proto_rawDescGZIP(), []int{2}
}

func (x *FailoverDestination) GetRef() *pbresource.Reference {
	if x != nil {
		return x.Ref
	}
	return nil
}

func (x *FailoverDestination) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *FailoverDestination) GetDatacenter() string {
	if x != nil {
		return x.Datacenter
	}
	return ""
}

var File_pbcatalog_v2beta1_failover_policy_proto protoreflect.FileDescriptor

var file_pbcatalog_v2beta1_failover_policy_proto_rawDesc = []byte{
	0x0a, 0x27, 0x70, 0x62, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2f, 0x76, 0x32, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2f, 0x66, 0x61, 0x69, 0x6c, 0x6f, 0x76, 0x65, 0x72, 0x5f, 0x70, 0x6f, 0x6c,
	0x69, 0x63, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x20, 0x68, 0x61, 0x73, 0x68, 0x69,
	0x63, 0x6f, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x2e, 0x63, 0x61, 0x74, 0x61,
	0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x32, 0x62, 0x65, 0x74, 0x61, 0x31, 0x1a, 0x1c, 0x70, 0x62, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x70, 0x62, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xba, 0x02, 0x0a, 0x0e, 0x46, 0x61, 0x69, 0x6c, 0x6f, 0x76, 0x65,
	0x72, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x48, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x68, 0x61, 0x73, 0x68, 0x69, 0x63,
	0x6f, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x2e, 0x63, 0x61, 0x74, 0x61, 0x6c,
	0x6f, 0x67, 0x2e, 0x76, 0x32, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x46, 0x61, 0x69, 0x6c, 0x6f,
	0x76, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x64, 0x0a, 0x0c, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x41, 0x2e, 0x68, 0x61, 0x73, 0x68, 0x69, 0x63,
	0x6f, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x2e, 0x63, 0x61, 0x74, 0x61, 0x6c,
	0x6f, 0x67, 0x2e, 0x76, 0x32, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x46, 0x61, 0x69, 0x6c, 0x6f,
	0x76, 0x65, 0x72, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0b, 0x70, 0x6f, 0x72, 0x74,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x1a, 0x70, 0x0a, 0x10, 0x50, 0x6f, 0x72, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x46, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x68,
	0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x2e,
	0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x32, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e,
	0x46, 0x61, 0x69, 0x6c, 0x6f, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x3a, 0x06, 0xa2, 0x93, 0x04, 0x02, 0x08,
	0x03, 0x22, 0xf0, 0x01, 0x0a, 0x0e, 0x46, 0x61, 0x69, 0x6c, 0x6f, 0x76, 0x65, 0x72, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x59, 0x0a, 0x0c, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x68, 0x61, 0x73,
	0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x2e, 0x63, 0x61,
	0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x32, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x46, 0x61,
	0x69, 0x6c, 0x6f, 0x76, 0x65, 0x72, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x0c, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12,
	0x42, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2e, 0x2e,
	0x68, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c,
	0x2e, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x32, 0x62, 0x65, 0x74, 0x61, 0x31,
	0x2e, 0x46, 0x61, 0x69, 0x6c, 0x6f, 0x76, 0x65, 0x72, 0x4d, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x6d,
	0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x25, 0x0a,
	0x0e, 0x73, 0x61, 0x6d, 0x65, 0x6e, 0x65, 0x73, 0x73, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x61, 0x6d, 0x65, 0x6e, 0x65, 0x73, 0x73, 0x47,
	0x72, 0x6f, 0x75, 0x70, 0x22, 0x81, 0x01, 0x0a, 0x13, 0x46, 0x61, 0x69, 0x6c, 0x6f, 0x76, 0x65,
	0x72, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x03,
	0x72, 0x65, 0x66, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x68, 0x61, 0x73, 0x68,
	0x69, 0x63, 0x6f, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x2e, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52,
	0x03, 0x72, 0x65, 0x66, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x64, 0x61, 0x74, 0x61,
	0x63, 0x65, 0x6e, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x61,
	0x74, 0x61, 0x63, 0x65, 0x6e, 0x74, 0x65, 0x72, 0x2a, 0x70, 0x0a, 0x0c, 0x46, 0x61, 0x69, 0x6c,
	0x6f, 0x76, 0x65, 0x72, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x1d, 0x0a, 0x19, 0x46, 0x41, 0x49, 0x4c,
	0x4f, 0x56, 0x45, 0x52, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1c, 0x0a, 0x18, 0x46, 0x41, 0x49, 0x4c, 0x4f,
	0x56, 0x45, 0x52, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x53, 0x45, 0x51, 0x55, 0x45, 0x4e, 0x54,
	0x49, 0x41, 0x4c, 0x10, 0x01, 0x12, 0x23, 0x0a, 0x1f, 0x46, 0x41, 0x49, 0x4c, 0x4f, 0x56, 0x45,
	0x52, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x4f, 0x52, 0x44, 0x45, 0x52, 0x5f, 0x42, 0x59, 0x5f,
	0x4c, 0x4f, 0x43, 0x41, 0x4c, 0x49, 0x54, 0x59, 0x10, 0x02, 0x42, 0xa9, 0x02, 0x0a, 0x24, 0x63,
	0x6f, 0x6d, 0x2e, 0x68, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e,
	0x73, 0x75, 0x6c, 0x2e, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x32, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x42, 0x13, 0x46, 0x61, 0x69, 0x6c, 0x6f, 0x76, 0x65, 0x72, 0x50, 0x6f, 0x6c,
	0x69, 0x63, 0x79, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x49, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70,
	0x2f, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2d, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x2f, 0x70, 0x62, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2f, 0x76,
	0x32, 0x62, 0x65, 0x74, 0x61, 0x31, 0x3b, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x76, 0x32,
	0x62, 0x65, 0x74, 0x61, 0x31, 0xa2, 0x02, 0x03, 0x48, 0x43, 0x43, 0xaa, 0x02, 0x20, 0x48, 0x61,
	0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x2e, 0x43, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x2e, 0x43,
	0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2e, 0x56, 0x32, 0x62, 0x65, 0x74, 0x61, 0x31, 0xca, 0x02,
	0x20, 0x48, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x5c, 0x43, 0x6f, 0x6e, 0x73, 0x75,
	0x6c, 0x5c, 0x43, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x5c, 0x56, 0x32, 0x62, 0x65, 0x74, 0x61,
	0x31, 0xe2, 0x02, 0x2c, 0x48, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x5c, 0x43, 0x6f,
	0x6e, 0x73, 0x75, 0x6c, 0x5c, 0x43, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x5c, 0x56, 0x32, 0x62,
	0x65, 0x74, 0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x23, 0x48, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x3a, 0x3a, 0x43, 0x6f,
	0x6e, 0x73, 0x75, 0x6c, 0x3a, 0x3a, 0x43, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x3a, 0x3a, 0x56,
	0x32, 0x62, 0x65, 0x74, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pbcatalog_v2beta1_failover_policy_proto_rawDescOnce sync.Once
	file_pbcatalog_v2beta1_failover_policy_proto_rawDescData = file_pbcatalog_v2beta1_failover_policy_proto_rawDesc
)

func file_pbcatalog_v2beta1_failover_policy_proto_rawDescGZIP() []byte {
	file_pbcatalog_v2beta1_failover_policy_proto_rawDescOnce.Do(func() {
		file_pbcatalog_v2beta1_failover_policy_proto_rawDescData = protoimpl.X.CompressGZIP(file_pbcatalog_v2beta1_failover_policy_proto_rawDescData)
	})
	return file_pbcatalog_v2beta1_failover_policy_proto_rawDescData
}

var file_pbcatalog_v2beta1_failover_policy_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pbcatalog_v2beta1_failover_policy_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_pbcatalog_v2beta1_failover_policy_proto_goTypes = []interface{}{
	(FailoverMode)(0),            // 0: hashicorp.consul.catalog.v2beta1.FailoverMode
	(*FailoverPolicy)(nil),       // 1: hashicorp.consul.catalog.v2beta1.FailoverPolicy
	(*FailoverConfig)(nil),       // 2: hashicorp.consul.catalog.v2beta1.FailoverConfig
	(*FailoverDestination)(nil),  // 3: hashicorp.consul.catalog.v2beta1.FailoverDestination
	nil,                          // 4: hashicorp.consul.catalog.v2beta1.FailoverPolicy.PortConfigsEntry
	(*pbresource.Reference)(nil), // 5: hashicorp.consul.resource.Reference
}
var file_pbcatalog_v2beta1_failover_policy_proto_depIdxs = []int32{
	2, // 0: hashicorp.consul.catalog.v2beta1.FailoverPolicy.config:type_name -> hashicorp.consul.catalog.v2beta1.FailoverConfig
	4, // 1: hashicorp.consul.catalog.v2beta1.FailoverPolicy.port_configs:type_name -> hashicorp.consul.catalog.v2beta1.FailoverPolicy.PortConfigsEntry
	3, // 2: hashicorp.consul.catalog.v2beta1.FailoverConfig.destinations:type_name -> hashicorp.consul.catalog.v2beta1.FailoverDestination
	0, // 3: hashicorp.consul.catalog.v2beta1.FailoverConfig.mode:type_name -> hashicorp.consul.catalog.v2beta1.FailoverMode
	5, // 4: hashicorp.consul.catalog.v2beta1.FailoverDestination.ref:type_name -> hashicorp.consul.resource.Reference
	2, // 5: hashicorp.consul.catalog.v2beta1.FailoverPolicy.PortConfigsEntry.value:type_name -> hashicorp.consul.catalog.v2beta1.FailoverConfig
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_pbcatalog_v2beta1_failover_policy_proto_init() }
func file_pbcatalog_v2beta1_failover_policy_proto_init() {
	if File_pbcatalog_v2beta1_failover_policy_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FailoverPolicy); i {
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
		file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FailoverConfig); i {
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
		file_pbcatalog_v2beta1_failover_policy_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FailoverDestination); i {
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
			RawDescriptor: file_pbcatalog_v2beta1_failover_policy_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pbcatalog_v2beta1_failover_policy_proto_goTypes,
		DependencyIndexes: file_pbcatalog_v2beta1_failover_policy_proto_depIdxs,
		EnumInfos:         file_pbcatalog_v2beta1_failover_policy_proto_enumTypes,
		MessageInfos:      file_pbcatalog_v2beta1_failover_policy_proto_msgTypes,
	}.Build()
	File_pbcatalog_v2beta1_failover_policy_proto = out.File
	file_pbcatalog_v2beta1_failover_policy_proto_rawDesc = nil
	file_pbcatalog_v2beta1_failover_policy_proto_goTypes = nil
	file_pbcatalog_v2beta1_failover_policy_proto_depIdxs = nil
}
