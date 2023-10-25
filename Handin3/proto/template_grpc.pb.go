// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.3
// source: proto/template.proto

package proto

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

const (
	Chittychat_Send_FullMethodName = "/proto.Chittychat/Send"
)

// ChittychatClient is the client API for Chittychat service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChittychatClient interface {
	Send(ctx context.Context, opts ...grpc.CallOption) (Chittychat_SendClient, error)
}

type chittychatClient struct {
	cc grpc.ClientConnInterface
}

func NewChittychatClient(cc grpc.ClientConnInterface) ChittychatClient {
	return &chittychatClient{cc}
}

func (c *chittychatClient) Send(ctx context.Context, opts ...grpc.CallOption) (Chittychat_SendClient, error) {
	stream, err := c.cc.NewStream(ctx, &Chittychat_ServiceDesc.Streams[0], Chittychat_Send_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &chittychatSendClient{stream}
	return x, nil
}

type Chittychat_SendClient interface {
	Send(*Message) error
	Recv() (*Verification, error)
	grpc.ClientStream
}

type chittychatSendClient struct {
	grpc.ClientStream
}

func (x *chittychatSendClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *chittychatSendClient) Recv() (*Verification, error) {
	m := new(Verification)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChittychatServer is the server API for Chittychat service.
// All implementations must embed UnimplementedChittychatServer
// for forward compatibility
type ChittychatServer interface {
	Send(Chittychat_SendServer) error
	mustEmbedUnimplementedChittychatServer()
}

// UnimplementedChittychatServer must be embedded to have forward compatible implementations.
type UnimplementedChittychatServer struct {
}

func (UnimplementedChittychatServer) Send(Chittychat_SendServer) error {
	return status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedChittychatServer) mustEmbedUnimplementedChittychatServer() {}

// UnsafeChittychatServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChittychatServer will
// result in compilation errors.
type UnsafeChittychatServer interface {
	mustEmbedUnimplementedChittychatServer()
}

func RegisterChittychatServer(s grpc.ServiceRegistrar, srv ChittychatServer) {
	s.RegisterService(&Chittychat_ServiceDesc, srv)
}

func _Chittychat_Send_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChittychatServer).Send(&chittychatSendServer{stream})
}

type Chittychat_SendServer interface {
	Send(*Verification) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type chittychatSendServer struct {
	grpc.ServerStream
}

func (x *chittychatSendServer) Send(m *Verification) error {
	return x.ServerStream.SendMsg(m)
}

func (x *chittychatSendServer) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Chittychat_ServiceDesc is the grpc.ServiceDesc for Chittychat service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Chittychat_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Chittychat",
	HandlerType: (*ChittychatServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Send",
			Handler:       _Chittychat_Send_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/template.proto",
}
