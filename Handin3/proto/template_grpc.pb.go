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
	Chittychat_Increment_FullMethodName       = "/proto.Chittychat/Increment"
	Chittychat_Send_FullMethodName            = "/proto.Chittychat/Send"
	Chittychat_SendChatMessage_FullMethodName = "/proto.Chittychat/SendChatMessage"
	Chittychat_HandleNewClient_FullMethodName = "/proto.Chittychat/HandleNewClient"
	Chittychat_ChatStream_FullMethodName      = "/proto.Chittychat/ChatStream"
)

// ChittychatClient is the client API for Chittychat service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChittychatClient interface {
	Increment(ctx context.Context, in *Amount, opts ...grpc.CallOption) (*Ack, error)
	Send(ctx context.Context, opts ...grpc.CallOption) (Chittychat_SendClient, error)
	SendChatMessage(ctx context.Context, in *ChatMessage, opts ...grpc.CallOption) (*Ack1, error)
	HandleNewClient(ctx context.Context, in *JoinMessage, opts ...grpc.CallOption) (*Ack1, error)
	ChatStream(ctx context.Context, opts ...grpc.CallOption) (Chittychat_ChatStreamClient, error)
}

type chittychatClient struct {
	cc grpc.ClientConnInterface
}

func NewChittychatClient(cc grpc.ClientConnInterface) ChittychatClient {
	return &chittychatClient{cc}
}

func (c *chittychatClient) Increment(ctx context.Context, in *Amount, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := c.cc.Invoke(ctx, Chittychat_Increment_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
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
	CloseAndRecv() (*Value, error)
	grpc.ClientStream
}

type chittychatSendClient struct {
	grpc.ClientStream
}

func (x *chittychatSendClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *chittychatSendClient) CloseAndRecv() (*Value, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Value)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *chittychatClient) SendChatMessage(ctx context.Context, in *ChatMessage, opts ...grpc.CallOption) (*Ack1, error) {
	out := new(Ack1)
	err := c.cc.Invoke(ctx, Chittychat_SendChatMessage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chittychatClient) HandleNewClient(ctx context.Context, in *JoinMessage, opts ...grpc.CallOption) (*Ack1, error) {
	out := new(Ack1)
	err := c.cc.Invoke(ctx, Chittychat_HandleNewClient_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chittychatClient) ChatStream(ctx context.Context, opts ...grpc.CallOption) (Chittychat_ChatStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Chittychat_ServiceDesc.Streams[1], Chittychat_ChatStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &chittychatChatStreamClient{stream}
	return x, nil
}

type Chittychat_ChatStreamClient interface {
	Send(*ChatMessage) error
	Recv() (*Ack1, error)
	grpc.ClientStream
}

type chittychatChatStreamClient struct {
	grpc.ClientStream
}

func (x *chittychatChatStreamClient) Send(m *ChatMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *chittychatChatStreamClient) Recv() (*Ack1, error) {
	m := new(Ack1)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChittychatServer is the server API for Chittychat service.
// All implementations must embed UnimplementedChittychatServer
// for forward compatibility
type ChittychatServer interface {
	Increment(context.Context, *Amount) (*Ack, error)
	Send(Chittychat_SendServer) error
	SendChatMessage(context.Context, *ChatMessage) (*Ack1, error)
	HandleNewClient(context.Context, *JoinMessage) (*Ack1, error)
	ChatStream(Chittychat_ChatStreamServer) error
	mustEmbedUnimplementedChittychatServer()
}

// UnimplementedChittychatServer must be embedded to have forward compatible implementations.
type UnimplementedChittychatServer struct {
}

func (UnimplementedChittychatServer) Increment(context.Context, *Amount) (*Ack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Increment not implemented")
}
func (UnimplementedChittychatServer) Send(Chittychat_SendServer) error {
	return status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedChittychatServer) SendChatMessage(context.Context, *ChatMessage) (*Ack1, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendChatMessage not implemented")
}
func (UnimplementedChittychatServer) HandleNewClient(context.Context, *JoinMessage) (*Ack1, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleNewClient not implemented")
}
func (UnimplementedChittychatServer) ChatStream(Chittychat_ChatStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method ChatStream not implemented")
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

func _Chittychat_Increment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Amount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChittychatServer).Increment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Chittychat_Increment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChittychatServer).Increment(ctx, req.(*Amount))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chittychat_Send_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChittychatServer).Send(&chittychatSendServer{stream})
}

type Chittychat_SendServer interface {
	SendAndClose(*Value) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type chittychatSendServer struct {
	grpc.ServerStream
}

func (x *chittychatSendServer) SendAndClose(m *Value) error {
	return x.ServerStream.SendMsg(m)
}

func (x *chittychatSendServer) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Chittychat_SendChatMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChittychatServer).SendChatMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Chittychat_SendChatMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChittychatServer).SendChatMessage(ctx, req.(*ChatMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chittychat_HandleNewClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChittychatServer).HandleNewClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Chittychat_HandleNewClient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChittychatServer).HandleNewClient(ctx, req.(*JoinMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chittychat_ChatStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChittychatServer).ChatStream(&chittychatChatStreamServer{stream})
}

type Chittychat_ChatStreamServer interface {
	Send(*Ack1) error
	Recv() (*ChatMessage, error)
	grpc.ServerStream
}

type chittychatChatStreamServer struct {
	grpc.ServerStream
}

func (x *chittychatChatStreamServer) Send(m *Ack1) error {
	return x.ServerStream.SendMsg(m)
}

func (x *chittychatChatStreamServer) Recv() (*ChatMessage, error) {
	m := new(ChatMessage)
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
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Increment",
			Handler:    _Chittychat_Increment_Handler,
		},
		{
			MethodName: "SendChatMessage",
			Handler:    _Chittychat_SendChatMessage_Handler,
		},
		{
			MethodName: "HandleNewClient",
			Handler:    _Chittychat_HandleNewClient_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Send",
			Handler:       _Chittychat_Send_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "ChatStream",
			Handler:       _Chittychat_ChatStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/template.proto",
}
