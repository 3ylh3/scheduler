// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: apiserver/qryexerec/query_execution_record.proto

package qryexerec

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// QueryExecutionRecordClient is the client API for QueryExecutionRecord service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QueryExecutionRecordClient interface {
	QueryExecutionRecord(ctx context.Context, in *QueryExecutionRecordRequest, opts ...grpc.CallOption) (*QueryExecutionRecordResponse, error)
}

type queryExecutionRecordClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryExecutionRecordClient(cc grpc.ClientConnInterface) QueryExecutionRecordClient {
	return &queryExecutionRecordClient{cc}
}

func (c *queryExecutionRecordClient) QueryExecutionRecord(ctx context.Context, in *QueryExecutionRecordRequest, opts ...grpc.CallOption) (*QueryExecutionRecordResponse, error) {
	out := new(QueryExecutionRecordResponse)
	err := c.cc.Invoke(ctx, "/qryexerec.QueryExecutionRecord/QueryExecutionRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryExecutionRecordServer is the server API for QueryExecutionRecord service.
// All implementations must embed UnimplementedQueryExecutionRecordServer
// for forward compatibility
type QueryExecutionRecordServer interface {
	QueryExecutionRecord(context.Context, *QueryExecutionRecordRequest) (*QueryExecutionRecordResponse, error)
	mustEmbedUnimplementedQueryExecutionRecordServer()
}

// UnimplementedQueryExecutionRecordServer must be embedded to have forward compatible implementations.
type UnimplementedQueryExecutionRecordServer struct {
}

func (UnimplementedQueryExecutionRecordServer) QueryExecutionRecord(context.Context, *QueryExecutionRecordRequest) (*QueryExecutionRecordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryExecutionRecord not implemented")
}
func (UnimplementedQueryExecutionRecordServer) mustEmbedUnimplementedQueryExecutionRecordServer() {}

// UnsafeQueryExecutionRecordServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QueryExecutionRecordServer will
// result in compilation errors.
type UnsafeQueryExecutionRecordServer interface {
	mustEmbedUnimplementedQueryExecutionRecordServer()
}

func RegisterQueryExecutionRecordServer(s grpc.ServiceRegistrar, srv QueryExecutionRecordServer) {
	s.RegisterService(&QueryExecutionRecord_ServiceDesc, srv)
}

func _QueryExecutionRecord_QueryExecutionRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryExecutionRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryExecutionRecordServer).QueryExecutionRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/qryexerec.QueryExecutionRecord/QueryExecutionRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryExecutionRecordServer).QueryExecutionRecord(ctx, req.(*QueryExecutionRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// QueryExecutionRecord_ServiceDesc is the grpc.ServiceDesc for QueryExecutionRecord service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var QueryExecutionRecord_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "qryexerec.QueryExecutionRecord",
	HandlerType: (*QueryExecutionRecordServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "QueryExecutionRecord",
			Handler:    _QueryExecutionRecord_QueryExecutionRecord_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "apiserver/qryexerec/query_execution_record.proto",
}
