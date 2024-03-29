// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: apiserver/deletejob/delete_job.proto

package deletejob

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

// DeleteJobClient is the client API for DeleteJob service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DeleteJobClient interface {
	DeleteJob(ctx context.Context, in *DeleteJobRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type deleteJobClient struct {
	cc grpc.ClientConnInterface
}

func NewDeleteJobClient(cc grpc.ClientConnInterface) DeleteJobClient {
	return &deleteJobClient{cc}
}

func (c *deleteJobClient) DeleteJob(ctx context.Context, in *DeleteJobRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/deletejob.DeleteJob/DeleteJob", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteJobServer is the server API for DeleteJob service.
// All implementations must embed UnimplementedDeleteJobServer
// for forward compatibility
type DeleteJobServer interface {
	DeleteJob(context.Context, *DeleteJobRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedDeleteJobServer()
}

// UnimplementedDeleteJobServer must be embedded to have forward compatible implementations.
type UnimplementedDeleteJobServer struct {
}

func (UnimplementedDeleteJobServer) DeleteJob(context.Context, *DeleteJobRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteJob not implemented")
}
func (UnimplementedDeleteJobServer) mustEmbedUnimplementedDeleteJobServer() {}

// UnsafeDeleteJobServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DeleteJobServer will
// result in compilation errors.
type UnsafeDeleteJobServer interface {
	mustEmbedUnimplementedDeleteJobServer()
}

func RegisterDeleteJobServer(s grpc.ServiceRegistrar, srv DeleteJobServer) {
	s.RegisterService(&DeleteJob_ServiceDesc, srv)
}

func _DeleteJob_DeleteJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeleteJobServer).DeleteJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/deletejob.DeleteJob/DeleteJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeleteJobServer).DeleteJob(ctx, req.(*DeleteJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DeleteJob_ServiceDesc is the grpc.ServiceDesc for DeleteJob service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DeleteJob_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "deletejob.DeleteJob",
	HandlerType: (*DeleteJobServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DeleteJob",
			Handler:    _DeleteJob_DeleteJob_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "apiserver/deletejob/delete_job.proto",
}
