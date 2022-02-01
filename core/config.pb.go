// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: core/config.proto

package core

import (
	serial "github.com/vmessocket/vmessocket/common/serial"
	transport "github.com/vmessocket/vmessocket/transport"
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

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Inbound  []*InboundHandlerConfig  `protobuf:"bytes,1,rep,name=inbound,proto3" json:"inbound,omitempty"`
	Outbound []*OutboundHandlerConfig `protobuf:"bytes,2,rep,name=outbound,proto3" json:"outbound,omitempty"`
	App      []*serial.TypedMessage   `protobuf:"bytes,4,rep,name=app,proto3" json:"app,omitempty"`
	// Deprecated: Do not use.
	Transport *transport.Config      `protobuf:"bytes,5,opt,name=transport,proto3" json:"transport,omitempty"`
	Extension []*serial.TypedMessage `protobuf:"bytes,6,rep,name=extension,proto3" json:"extension,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_core_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_core_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetInbound() []*InboundHandlerConfig {
	if x != nil {
		return x.Inbound
	}
	return nil
}

func (x *Config) GetOutbound() []*OutboundHandlerConfig {
	if x != nil {
		return x.Outbound
	}
	return nil
}

func (x *Config) GetApp() []*serial.TypedMessage {
	if x != nil {
		return x.App
	}
	return nil
}

// Deprecated: Do not use.
func (x *Config) GetTransport() *transport.Config {
	if x != nil {
		return x.Transport
	}
	return nil
}

func (x *Config) GetExtension() []*serial.TypedMessage {
	if x != nil {
		return x.Extension
	}
	return nil
}

type InboundHandlerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tag              string               `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	ReceiverSettings *serial.TypedMessage `protobuf:"bytes,2,opt,name=receiver_settings,json=receiverSettings,proto3" json:"receiver_settings,omitempty"`
	ProxySettings    *serial.TypedMessage `protobuf:"bytes,3,opt,name=proxy_settings,json=proxySettings,proto3" json:"proxy_settings,omitempty"`
}

func (x *InboundHandlerConfig) Reset() {
	*x = InboundHandlerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InboundHandlerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InboundHandlerConfig) ProtoMessage() {}

func (x *InboundHandlerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_core_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InboundHandlerConfig.ProtoReflect.Descriptor instead.
func (*InboundHandlerConfig) Descriptor() ([]byte, []int) {
	return file_core_config_proto_rawDescGZIP(), []int{1}
}

func (x *InboundHandlerConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *InboundHandlerConfig) GetReceiverSettings() *serial.TypedMessage {
	if x != nil {
		return x.ReceiverSettings
	}
	return nil
}

func (x *InboundHandlerConfig) GetProxySettings() *serial.TypedMessage {
	if x != nil {
		return x.ProxySettings
	}
	return nil
}

type OutboundHandlerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tag            string               `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	SenderSettings *serial.TypedMessage `protobuf:"bytes,2,opt,name=sender_settings,json=senderSettings,proto3" json:"sender_settings,omitempty"`
	ProxySettings  *serial.TypedMessage `protobuf:"bytes,3,opt,name=proxy_settings,json=proxySettings,proto3" json:"proxy_settings,omitempty"`
	Expire         int64                `protobuf:"varint,4,opt,name=expire,proto3" json:"expire,omitempty"`
	Comment        string               `protobuf:"bytes,5,opt,name=comment,proto3" json:"comment,omitempty"`
}

func (x *OutboundHandlerConfig) Reset() {
	*x = OutboundHandlerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_config_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OutboundHandlerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OutboundHandlerConfig) ProtoMessage() {}

func (x *OutboundHandlerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_core_config_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OutboundHandlerConfig.ProtoReflect.Descriptor instead.
func (*OutboundHandlerConfig) Descriptor() ([]byte, []int) {
	return file_core_config_proto_rawDescGZIP(), []int{2}
}

func (x *OutboundHandlerConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *OutboundHandlerConfig) GetSenderSettings() *serial.TypedMessage {
	if x != nil {
		return x.SenderSettings
	}
	return nil
}

func (x *OutboundHandlerConfig) GetProxySettings() *serial.TypedMessage {
	if x != nil {
		return x.ProxySettings
	}
	return nil
}

func (x *OutboundHandlerConfig) GetExpire() int64 {
	if x != nil {
		return x.Expire
	}
	return 0
}

func (x *OutboundHandlerConfig) GetComment() string {
	if x != nil {
		return x.Comment
	}
	return ""
}

var File_core_config_proto protoreflect.FileDescriptor

var file_core_config_proto_rawDesc = []byte{
	0x0a, 0x11, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e,
	0x63, 0x6f, 0x72, 0x65, 0x1a, 0x21, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x73, 0x65, 0x72,
	0x69, 0x61, 0x6c, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x64, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f,
	0x72, 0x74, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xe2, 0x02, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x3f, 0x0a, 0x07, 0x69, 0x6e,
	0x62, 0x6f, 0x75, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x76, 0x6d,
	0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x49, 0x6e,
	0x62, 0x6f, 0x75, 0x6e, 0x64, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x52, 0x07, 0x69, 0x6e, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x42, 0x0a, 0x08, 0x6f,
	0x75, 0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e,
	0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x4f, 0x75, 0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x08, 0x6f, 0x75, 0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x12,
	0x3d, 0x0a, 0x03, 0x61, 0x70, 0x70, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x76,
	0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x2e, 0x54, 0x79, 0x70,
	0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x03, 0x61, 0x70, 0x70, 0x12, 0x43,
	0x0a, 0x09, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x21, 0x2e, 0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x42, 0x02, 0x18, 0x01, 0x52, 0x09, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x12, 0x49, 0x0a, 0x09, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63,
	0x6b, 0x65, 0x74, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x52, 0x09, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x4a, 0x04,
	0x08, 0x03, 0x10, 0x04, 0x22, 0xd6, 0x01, 0x0a, 0x14, 0x49, 0x6e, 0x62, 0x6f, 0x75, 0x6e, 0x64,
	0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x10, 0x0a,
	0x03, 0x74, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x61, 0x67, 0x12,
	0x58, 0x0a, 0x11, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x5f, 0x73, 0x65, 0x74, 0x74,
	0x69, 0x6e, 0x67, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x76, 0x6d, 0x65,
	0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x64,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x10, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65,
	0x72, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x12, 0x52, 0x0a, 0x0e, 0x70, 0x72, 0x6f,
	0x78, 0x79, 0x5f, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x2b, 0x2e, 0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x69, 0x61,
	0x6c, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x0d,
	0x70, 0x72, 0x6f, 0x78, 0x79, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x22, 0x85, 0x02,
	0x0a, 0x15, 0x4f, 0x75, 0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65,
	0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x61, 0x67, 0x12, 0x54, 0x0a, 0x0f, 0x73, 0x65, 0x6e,
	0x64, 0x65, 0x72, 0x5f, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e,
	0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x69,
	0x61, 0x6c, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52,
	0x0e, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x12,
	0x52, 0x0a, 0x0e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f,
	0x63, 0x6b, 0x65, 0x74, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x64, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x52, 0x0d, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x53, 0x65, 0x74, 0x74, 0x69,
	0x6e, 0x67, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x06, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x42, 0x50, 0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x6d, 0x65,
	0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x50, 0x01, 0x5a, 0x25,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x65, 0x73, 0x73,
	0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2f, 0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74,
	0x2f, 0x63, 0x6f, 0x72, 0x65, 0xaa, 0x02, 0x0f, 0x76, 0x6d, 0x65, 0x73, 0x73, 0x6f, 0x63, 0x6b,
	0x65, 0x74, 0x2e, 0x43, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_core_config_proto_rawDescOnce sync.Once
	file_core_config_proto_rawDescData = file_core_config_proto_rawDesc
)

func file_core_config_proto_rawDescGZIP() []byte {
	file_core_config_proto_rawDescOnce.Do(func() {
		file_core_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_core_config_proto_rawDescData)
	})
	return file_core_config_proto_rawDescData
}

var file_core_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_core_config_proto_goTypes = []interface{}{
	(*Config)(nil),                // 0: vmessocket.core.Config
	(*InboundHandlerConfig)(nil),  // 1: vmessocket.core.InboundHandlerConfig
	(*OutboundHandlerConfig)(nil), // 2: vmessocket.core.OutboundHandlerConfig
	(*serial.TypedMessage)(nil),   // 3: vmessocket.core.common.serial.TypedMessage
	(*transport.Config)(nil),      // 4: vmessocket.core.transport.Config
}
var file_core_config_proto_depIdxs = []int32{
	1, // 0: vmessocket.core.Config.inbound:type_name -> vmessocket.core.InboundHandlerConfig
	2, // 1: vmessocket.core.Config.outbound:type_name -> vmessocket.core.OutboundHandlerConfig
	3, // 2: vmessocket.core.Config.app:type_name -> vmessocket.core.common.serial.TypedMessage
	4, // 3: vmessocket.core.Config.transport:type_name -> vmessocket.core.transport.Config
	3, // 4: vmessocket.core.Config.extension:type_name -> vmessocket.core.common.serial.TypedMessage
	3, // 5: vmessocket.core.InboundHandlerConfig.receiver_settings:type_name -> vmessocket.core.common.serial.TypedMessage
	3, // 6: vmessocket.core.InboundHandlerConfig.proxy_settings:type_name -> vmessocket.core.common.serial.TypedMessage
	3, // 7: vmessocket.core.OutboundHandlerConfig.sender_settings:type_name -> vmessocket.core.common.serial.TypedMessage
	3, // 8: vmessocket.core.OutboundHandlerConfig.proxy_settings:type_name -> vmessocket.core.common.serial.TypedMessage
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_core_config_proto_init() }
func file_core_config_proto_init() {
	if File_core_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_core_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
		file_core_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InboundHandlerConfig); i {
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
		file_core_config_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OutboundHandlerConfig); i {
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
			RawDescriptor: file_core_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_core_config_proto_goTypes,
		DependencyIndexes: file_core_config_proto_depIdxs,
		MessageInfos:      file_core_config_proto_msgTypes,
	}.Build()
	File_core_config_proto = out.File
	file_core_config_proto_rawDesc = nil
	file_core_config_proto_goTypes = nil
	file_core_config_proto_depIdxs = nil
}
