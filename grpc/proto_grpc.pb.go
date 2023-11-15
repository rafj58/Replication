// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.4
// source: grpc/proto.proto

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
	AuctionService_Bid_FullMethodName          = "/proto.AuctionService/Bid"
	AuctionService_Result_FullMethodName       = "/proto.AuctionService/Result"
	AuctionService_AskForMaster_FullMethodName = "/proto.AuctionService/AskForMaster"
)

// AuctionServiceClient is the client API for AuctionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuctionServiceClient interface {
	Bid(ctx context.Context, in *Amount, opts ...grpc.CallOption) (*Acknowledge, error)
	Result(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Outcome, error)
	AskForMaster(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Peer, error)
}

type auctionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuctionServiceClient(cc grpc.ClientConnInterface) AuctionServiceClient {
	return &auctionServiceClient{cc}
}

func (c *auctionServiceClient) Bid(ctx context.Context, in *Amount, opts ...grpc.CallOption) (*Acknowledge, error) {
	out := new(Acknowledge)
	err := c.cc.Invoke(ctx, AuctionService_Bid_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionServiceClient) Result(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Outcome, error) {
	out := new(Outcome)
	err := c.cc.Invoke(ctx, AuctionService_Result_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionServiceClient) AskForMaster(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Peer, error) {
	out := new(Peer)
	err := c.cc.Invoke(ctx, AuctionService_AskForMaster_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuctionServiceServer is the server API for AuctionService service.
// All implementations must embed UnimplementedAuctionServiceServer
// for forward compatibility
type AuctionServiceServer interface {
	Bid(context.Context, *Amount) (*Acknowledge, error)
	Result(context.Context, *Empty) (*Outcome, error)
	AskForMaster(context.Context, *Empty) (*Peer, error)
	mustEmbedUnimplementedAuctionServiceServer()
}

// UnimplementedAuctionServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuctionServiceServer struct {
}

func (UnimplementedAuctionServiceServer) Bid(context.Context, *Amount) (*Acknowledge, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bid not implemented")
}
func (UnimplementedAuctionServiceServer) Result(context.Context, *Empty) (*Outcome, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Result not implemented")
}
func (UnimplementedAuctionServiceServer) AskForMaster(context.Context, *Empty) (*Peer, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AskForMaster not implemented")
}
func (UnimplementedAuctionServiceServer) mustEmbedUnimplementedAuctionServiceServer() {}

// UnsafeAuctionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuctionServiceServer will
// result in compilation errors.
type UnsafeAuctionServiceServer interface {
	mustEmbedUnimplementedAuctionServiceServer()
}

func RegisterAuctionServiceServer(s grpc.ServiceRegistrar, srv AuctionServiceServer) {
	s.RegisterService(&AuctionService_ServiceDesc, srv)
}

func _AuctionService_Bid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Amount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServiceServer).Bid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuctionService_Bid_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServiceServer).Bid(ctx, req.(*Amount))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuctionService_Result_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServiceServer).Result(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuctionService_Result_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServiceServer).Result(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuctionService_AskForMaster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServiceServer).AskForMaster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuctionService_AskForMaster_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServiceServer).AskForMaster(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// AuctionService_ServiceDesc is the grpc.ServiceDesc for AuctionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuctionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.AuctionService",
	HandlerType: (*AuctionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Bid",
			Handler:    _AuctionService_Bid_Handler,
		},
		{
			MethodName: "Result",
			Handler:    _AuctionService_Result_Handler,
		},
		{
			MethodName: "AskForMaster",
			Handler:    _AuctionService_AskForMaster_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/proto.proto",
}

const (
	DistributedService_Election_FullMethodName    = "/proto.DistributedService/Election"
	DistributedService_Coordinator_FullMethodName = "/proto.DistributedService/Coordinator"
)

// DistributedServiceClient is the client API for DistributedService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DistributedServiceClient interface {
	// bully algorithm
	Election(ctx context.Context, in *Peer, opts ...grpc.CallOption) (*Acknowledge, error)
	Coordinator(ctx context.Context, in *Peer, opts ...grpc.CallOption) (*Acknowledge, error)
}

type distributedServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDistributedServiceClient(cc grpc.ClientConnInterface) DistributedServiceClient {
	return &distributedServiceClient{cc}
}

func (c *distributedServiceClient) Election(ctx context.Context, in *Peer, opts ...grpc.CallOption) (*Acknowledge, error) {
	out := new(Acknowledge)
	err := c.cc.Invoke(ctx, DistributedService_Election_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *distributedServiceClient) Coordinator(ctx context.Context, in *Peer, opts ...grpc.CallOption) (*Acknowledge, error) {
	out := new(Acknowledge)
	err := c.cc.Invoke(ctx, DistributedService_Coordinator_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DistributedServiceServer is the server API for DistributedService service.
// All implementations must embed UnimplementedDistributedServiceServer
// for forward compatibility
type DistributedServiceServer interface {
	// bully algorithm
	Election(context.Context, *Peer) (*Acknowledge, error)
	Coordinator(context.Context, *Peer) (*Acknowledge, error)
	mustEmbedUnimplementedDistributedServiceServer()
}

// UnimplementedDistributedServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDistributedServiceServer struct {
}

func (UnimplementedDistributedServiceServer) Election(context.Context, *Peer) (*Acknowledge, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Election not implemented")
}
func (UnimplementedDistributedServiceServer) Coordinator(context.Context, *Peer) (*Acknowledge, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Coordinator not implemented")
}
func (UnimplementedDistributedServiceServer) mustEmbedUnimplementedDistributedServiceServer() {}

// UnsafeDistributedServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DistributedServiceServer will
// result in compilation errors.
type UnsafeDistributedServiceServer interface {
	mustEmbedUnimplementedDistributedServiceServer()
}

func RegisterDistributedServiceServer(s grpc.ServiceRegistrar, srv DistributedServiceServer) {
	s.RegisterService(&DistributedService_ServiceDesc, srv)
}

func _DistributedService_Election_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Peer)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DistributedServiceServer).Election(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DistributedService_Election_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DistributedServiceServer).Election(ctx, req.(*Peer))
	}
	return interceptor(ctx, in, info, handler)
}

func _DistributedService_Coordinator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Peer)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DistributedServiceServer).Coordinator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DistributedService_Coordinator_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DistributedServiceServer).Coordinator(ctx, req.(*Peer))
	}
	return interceptor(ctx, in, info, handler)
}

// DistributedService_ServiceDesc is the grpc.ServiceDesc for DistributedService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DistributedService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.DistributedService",
	HandlerType: (*DistributedServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Election",
			Handler:    _DistributedService_Election_Handler,
		},
		{
			MethodName: "Coordinator",
			Handler:    _DistributedService_Coordinator_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/proto.proto",
}
