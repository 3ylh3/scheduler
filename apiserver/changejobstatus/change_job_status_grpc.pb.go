// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: apiserver/changejobstatus/change_job_status.proto

package changejobstatus

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ChangeJobStatusClient is the client API for ChangeJobStatus service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChangeJobStatusClient interface {
	ChangeJobStatus(ctx context.Context, in *ChangeJobStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type changeJobStatusClient struct {
	cc grpc.ClientConnInterface
}

func NewChangeJobStatusClient(cc grpc.ClientConnInterface) ChangeJobStatusClient {
	return &changeJobStatusClient{cc}
}

func (c *changeJobStatusClient) ChangeJobStatus(ctx context.Context, in *ChangeJobStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/changejobstatus.ChangeJobStatus/ChangeJobStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChangeJobStatusServer is the server API for ChangeJobStatus service.
// All implementations must embed UnimplementedChangeJobStatusServer
// for forward compatibility
type ChangeJobStatusServer interface {
	ChangeJobStatus(context.Context, *ChangeJobStatusRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedChangeJobStatusServer()
}

// UnimplementedChangeJobStatusServer must be embedded to have forward compatible implementations.
type UnimplementedChangeJobStatusServer struct {
}

func (UnimplementedChangeJobStatusServer) ChangeJobStatus(context.Context, *ChangeJobStatusRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeJobStatus not implemented")
}
func (UnimplementedChangeJobStatusServer) mustEmbedUnimplementedChangeJobStatusServer() {}

// UnsafeChangeJobStatusServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChangeJobStatusServer will
// result in compilation errors.
type UnsafeChangeJobStatusServer interface {
	mustEmbedUnimplementedChangeJobStatusServer()
}

func RegisterChangeJobStatusServer(s grpc.ServiceRegistrar, srv ChangeJobStatusServer) {
	s.RegisterService(&ChangeJobStatus_ServiceDesc, srv)
}

func _ChangeJobStatus_ChangeJobStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangeJobStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChangeJobStatusServer).ChangeJobStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/changejobstatus.ChangeJobStatus/ChangeJobStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChangeJobStatusServer).ChangeJobStatus(ctx, req.(*ChangeJobStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ChangeJobStatus_ServiceDesc is the grpc.ServiceDesc for ChangeJobStatus service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ChangeJobStatus_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "changejobstatus.ChangeJobStatus",
	HandlerType: (*ChangeJobStatusServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ChangeJobStatus",
			Handler:    _ChangeJobStatus_ChangeJobStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "apiserver/changejobstatus/change_job_status.proto",
}