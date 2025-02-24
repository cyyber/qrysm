// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: proto/zond/service/node_service.proto

package service

import (
	context "context"
	reflect "reflect"

	v1 "github.com/theQRL/qrysm/proto/zond/v1"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_proto_zond_service_node_service_proto protoreflect.FileDescriptor

var file_proto_zond_service_node_service_proto_rawDesc = []byte{
	0x0a, 0x25, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x7a, 0x6f, 0x6e, 0x64, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x74, 0x68, 0x65, 0x71, 0x72, 0x6c, 0x2e,
	0x7a, 0x6f, 0x6e, 0x64, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x1a, 0x1c, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d,
	0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x7a, 0x6f, 0x6e, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x32, 0xaa, 0x05, 0x0a, 0x0a, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x4e, 0x6f,
	0x64, 0x65, 0x12, 0x70, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x20, 0x2e, 0x74, 0x68, 0x65, 0x71,
	0x72, 0x6c, 0x2e, 0x7a, 0x6f, 0x6e, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x27, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x21, 0x12, 0x1f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x7a,
	0x6f, 0x6e, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x2f, 0x69, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x12, 0x6e, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x50, 0x65, 0x65, 0x72,
	0x73, 0x12, 0x1c, 0x2e, 0x74, 0x68, 0x65, 0x71, 0x72, 0x6c, 0x2e, 0x7a, 0x6f, 0x6e, 0x64, 0x2e,
	0x76, 0x31, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1d, 0x2e, 0x74, 0x68, 0x65, 0x71, 0x72, 0x6c, 0x2e, 0x7a, 0x6f, 0x6e, 0x64, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x65, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x24,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1e, 0x12, 0x1c, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x7a, 0x6f, 0x6e, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x2f, 0x70,
	0x65, 0x65, 0x72, 0x73, 0x12, 0x74, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x50, 0x65, 0x65, 0x72, 0x12,
	0x1b, 0x2e, 0x74, 0x68, 0x65, 0x71, 0x72, 0x6c, 0x2e, 0x7a, 0x6f, 0x6e, 0x64, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x65, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x74,
	0x68, 0x65, 0x71, 0x72, 0x6c, 0x2e, 0x7a, 0x6f, 0x6e, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x65,
	0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2e, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x28, 0x12, 0x26, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x7a, 0x6f,
	0x6e, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x2f, 0x70, 0x65, 0x65, 0x72, 0x73,
	0x2f, 0x7b, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x7d, 0x12, 0x71, 0x0a, 0x09, 0x50, 0x65,
	0x65, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a,
	0x21, 0x2e, 0x74, 0x68, 0x65, 0x71, 0x72, 0x6c, 0x2e, 0x7a, 0x6f, 0x6e, 0x64, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x65, 0x65, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x29, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x23, 0x12, 0x21, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x7a, 0x6f, 0x6e, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f,
	0x64, 0x65, 0x2f, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x6d, 0x0a,
	0x0a, 0x47, 0x65, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x1a, 0x1f, 0x2e, 0x74, 0x68, 0x65, 0x71, 0x72, 0x6c, 0x2e, 0x7a, 0x6f, 0x6e,
	0x64, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x26, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x20, 0x12, 0x1e, 0x2f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x7a, 0x6f, 0x6e, 0x64, 0x2f, 0x76, 0x31, 0x2f,
	0x6e, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x62, 0x0a, 0x09,
	0x47, 0x65, 0x74, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x25, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x1f, 0x12, 0x1d, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x7a, 0x6f, 0x6e,
	0x64, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x2f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68,
	0x42, 0x85, 0x01, 0x0a, 0x17, 0x6f, 0x72, 0x67, 0x2e, 0x74, 0x68, 0x65, 0x71, 0x72, 0x6c, 0x2e,
	0x7a, 0x6f, 0x6e, 0x64, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x10, 0x4e, 0x6f,
	0x64, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x68, 0x65,
	0x51, 0x52, 0x4c, 0x2f, 0x71, 0x72, 0x79, 0x73, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x7a, 0x6f, 0x6e, 0x64, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0xaa, 0x02, 0x13, 0x54,
	0x68, 0x65, 0x51, 0x52, 0x4c, 0x2e, 0x5a, 0x6f, 0x6e, 0x64, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0xca, 0x02, 0x13, 0x54, 0x68, 0x65, 0x51, 0x52, 0x4c, 0x5c, 0x5a, 0x6f, 0x6e, 0x64,
	0x5c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_proto_zond_service_node_service_proto_goTypes = []interface{}{
	(*emptypb.Empty)(nil),        // 0: google.protobuf.Empty
	(*v1.PeersRequest)(nil),      // 1: theqrl.zond.v1.PeersRequest
	(*v1.PeerRequest)(nil),       // 2: theqrl.zond.v1.PeerRequest
	(*v1.IdentityResponse)(nil),  // 3: theqrl.zond.v1.IdentityResponse
	(*v1.PeersResponse)(nil),     // 4: theqrl.zond.v1.PeersResponse
	(*v1.PeerResponse)(nil),      // 5: theqrl.zond.v1.PeerResponse
	(*v1.PeerCountResponse)(nil), // 6: theqrl.zond.v1.PeerCountResponse
	(*v1.VersionResponse)(nil),   // 7: theqrl.zond.v1.VersionResponse
}
var file_proto_zond_service_node_service_proto_depIdxs = []int32{
	0, // 0: theqrl.zond.service.BeaconNode.GetIdentity:input_type -> google.protobuf.Empty
	1, // 1: theqrl.zond.service.BeaconNode.ListPeers:input_type -> theqrl.zond.v1.PeersRequest
	2, // 2: theqrl.zond.service.BeaconNode.GetPeer:input_type -> theqrl.zond.v1.PeerRequest
	0, // 3: theqrl.zond.service.BeaconNode.PeerCount:input_type -> google.protobuf.Empty
	0, // 4: theqrl.zond.service.BeaconNode.GetVersion:input_type -> google.protobuf.Empty
	0, // 5: theqrl.zond.service.BeaconNode.GetHealth:input_type -> google.protobuf.Empty
	3, // 6: theqrl.zond.service.BeaconNode.GetIdentity:output_type -> theqrl.zond.v1.IdentityResponse
	4, // 7: theqrl.zond.service.BeaconNode.ListPeers:output_type -> theqrl.zond.v1.PeersResponse
	5, // 8: theqrl.zond.service.BeaconNode.GetPeer:output_type -> theqrl.zond.v1.PeerResponse
	6, // 9: theqrl.zond.service.BeaconNode.PeerCount:output_type -> theqrl.zond.v1.PeerCountResponse
	7, // 10: theqrl.zond.service.BeaconNode.GetVersion:output_type -> theqrl.zond.v1.VersionResponse
	0, // 11: theqrl.zond.service.BeaconNode.GetHealth:output_type -> google.protobuf.Empty
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_zond_service_node_service_proto_init() }
func file_proto_zond_service_node_service_proto_init() {
	if File_proto_zond_service_node_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_zond_service_node_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_zond_service_node_service_proto_goTypes,
		DependencyIndexes: file_proto_zond_service_node_service_proto_depIdxs,
	}.Build()
	File_proto_zond_service_node_service_proto = out.File
	file_proto_zond_service_node_service_proto_rawDesc = nil
	file_proto_zond_service_node_service_proto_goTypes = nil
	file_proto_zond_service_node_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// BeaconNodeClient is the client API for BeaconNode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BeaconNodeClient interface {
	GetIdentity(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.IdentityResponse, error)
	ListPeers(ctx context.Context, in *v1.PeersRequest, opts ...grpc.CallOption) (*v1.PeersResponse, error)
	GetPeer(ctx context.Context, in *v1.PeerRequest, opts ...grpc.CallOption) (*v1.PeerResponse, error)
	PeerCount(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.PeerCountResponse, error)
	GetVersion(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.VersionResponse, error)
	GetHealth(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type beaconNodeClient struct {
	cc grpc.ClientConnInterface
}

func NewBeaconNodeClient(cc grpc.ClientConnInterface) BeaconNodeClient {
	return &beaconNodeClient{cc}
}

func (c *beaconNodeClient) GetIdentity(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.IdentityResponse, error) {
	out := new(v1.IdentityResponse)
	err := c.cc.Invoke(ctx, "/theqrl.zond.service.BeaconNode/GetIdentity", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconNodeClient) ListPeers(ctx context.Context, in *v1.PeersRequest, opts ...grpc.CallOption) (*v1.PeersResponse, error) {
	out := new(v1.PeersResponse)
	err := c.cc.Invoke(ctx, "/theqrl.zond.service.BeaconNode/ListPeers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconNodeClient) GetPeer(ctx context.Context, in *v1.PeerRequest, opts ...grpc.CallOption) (*v1.PeerResponse, error) {
	out := new(v1.PeerResponse)
	err := c.cc.Invoke(ctx, "/theqrl.zond.service.BeaconNode/GetPeer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconNodeClient) PeerCount(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.PeerCountResponse, error) {
	out := new(v1.PeerCountResponse)
	err := c.cc.Invoke(ctx, "/theqrl.zond.service.BeaconNode/PeerCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconNodeClient) GetVersion(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.VersionResponse, error) {
	out := new(v1.VersionResponse)
	err := c.cc.Invoke(ctx, "/theqrl.zond.service.BeaconNode/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconNodeClient) GetHealth(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/theqrl.zond.service.BeaconNode/GetHealth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BeaconNodeServer is the server API for BeaconNode service.
type BeaconNodeServer interface {
	GetIdentity(context.Context, *emptypb.Empty) (*v1.IdentityResponse, error)
	ListPeers(context.Context, *v1.PeersRequest) (*v1.PeersResponse, error)
	GetPeer(context.Context, *v1.PeerRequest) (*v1.PeerResponse, error)
	PeerCount(context.Context, *emptypb.Empty) (*v1.PeerCountResponse, error)
	GetVersion(context.Context, *emptypb.Empty) (*v1.VersionResponse, error)
	GetHealth(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
}

// UnimplementedBeaconNodeServer can be embedded to have forward compatible implementations.
type UnimplementedBeaconNodeServer struct {
}

func (*UnimplementedBeaconNodeServer) GetIdentity(context.Context, *emptypb.Empty) (*v1.IdentityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIdentity not implemented")
}
func (*UnimplementedBeaconNodeServer) ListPeers(context.Context, *v1.PeersRequest) (*v1.PeersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPeers not implemented")
}
func (*UnimplementedBeaconNodeServer) GetPeer(context.Context, *v1.PeerRequest) (*v1.PeerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPeer not implemented")
}
func (*UnimplementedBeaconNodeServer) PeerCount(context.Context, *emptypb.Empty) (*v1.PeerCountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PeerCount not implemented")
}
func (*UnimplementedBeaconNodeServer) GetVersion(context.Context, *emptypb.Empty) (*v1.VersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}
func (*UnimplementedBeaconNodeServer) GetHealth(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHealth not implemented")
}

func RegisterBeaconNodeServer(s *grpc.Server, srv BeaconNodeServer) {
	s.RegisterService(&_BeaconNode_serviceDesc, srv)
}

func _BeaconNode_GetIdentity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconNodeServer).GetIdentity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/theqrl.zond.service.BeaconNode/GetIdentity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconNodeServer).GetIdentity(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconNode_ListPeers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.PeersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconNodeServer).ListPeers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/theqrl.zond.service.BeaconNode/ListPeers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconNodeServer).ListPeers(ctx, req.(*v1.PeersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconNode_GetPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.PeerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconNodeServer).GetPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/theqrl.zond.service.BeaconNode/GetPeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconNodeServer).GetPeer(ctx, req.(*v1.PeerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconNode_PeerCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconNodeServer).PeerCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/theqrl.zond.service.BeaconNode/PeerCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconNodeServer).PeerCount(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconNode_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconNodeServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/theqrl.zond.service.BeaconNode/GetVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconNodeServer).GetVersion(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconNode_GetHealth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconNodeServer).GetHealth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/theqrl.zond.service.BeaconNode/GetHealth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconNodeServer).GetHealth(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _BeaconNode_serviceDesc = grpc.ServiceDesc{
	ServiceName: "theqrl.zond.service.BeaconNode",
	HandlerType: (*BeaconNodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetIdentity",
			Handler:    _BeaconNode_GetIdentity_Handler,
		},
		{
			MethodName: "ListPeers",
			Handler:    _BeaconNode_ListPeers_Handler,
		},
		{
			MethodName: "GetPeer",
			Handler:    _BeaconNode_GetPeer_Handler,
		},
		{
			MethodName: "PeerCount",
			Handler:    _BeaconNode_PeerCount_Handler,
		},
		{
			MethodName: "GetVersion",
			Handler:    _BeaconNode_GetVersion_Handler,
		},
		{
			MethodName: "GetHealth",
			Handler:    _BeaconNode_GetHealth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/zond/service/node_service.proto",
}
