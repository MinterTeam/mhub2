// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package tx_committer

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// TxCommitterClient is the client API for TxCommitter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TxCommitterClient interface {
	CommitTx(ctx context.Context, in *CommitTxRequest, opts ...grpc.CallOption) (*CommitTxReply, error)
	Address(ctx context.Context, in *AddressRequest, opts ...grpc.CallOption) (*AddressReply, error)
}

type txCommitterClient struct {
	cc grpc.ClientConnInterface
}

func NewTxCommitterClient(cc grpc.ClientConnInterface) TxCommitterClient {
	return &txCommitterClient{cc}
}

func (c *txCommitterClient) CommitTx(ctx context.Context, in *CommitTxRequest, opts ...grpc.CallOption) (*CommitTxReply, error) {
	out := new(CommitTxReply)
	err := c.cc.Invoke(ctx, "/tx_committer.TxCommitter/CommitTx", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txCommitterClient) Address(ctx context.Context, in *AddressRequest, opts ...grpc.CallOption) (*AddressReply, error) {
	out := new(AddressReply)
	err := c.cc.Invoke(ctx, "/tx_committer.TxCommitter/Address", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TxCommitterServer is the server API for TxCommitter service.
// All implementations must embed UnimplementedTxCommitterServer
// for forward compatibility
type TxCommitterServer interface {
	CommitTx(context.Context, *CommitTxRequest) (*CommitTxReply, error)
	Address(context.Context, *AddressRequest) (*AddressReply, error)
	mustEmbedUnimplementedTxCommitterServer()
}

// UnimplementedTxCommitterServer must be embedded to have forward compatible implementations.
type UnimplementedTxCommitterServer struct {
}

func (UnimplementedTxCommitterServer) CommitTx(context.Context, *CommitTxRequest) (*CommitTxReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitTx not implemented")
}
func (UnimplementedTxCommitterServer) Address(context.Context, *AddressRequest) (*AddressReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Address not implemented")
}
func (UnimplementedTxCommitterServer) mustEmbedUnimplementedTxCommitterServer() {}

// UnsafeTxCommitterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TxCommitterServer will
// result in compilation errors.
type UnsafeTxCommitterServer interface {
	mustEmbedUnimplementedTxCommitterServer()
}

func RegisterTxCommitterServer(s *grpc.Server, srv TxCommitterServer) {
	s.RegisterService(&_TxCommitter_serviceDesc, srv)
}

func _TxCommitter_CommitTx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitTxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxCommitterServer).CommitTx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tx_committer.TxCommitter/CommitTx",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxCommitterServer).CommitTx(ctx, req.(*CommitTxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxCommitter_Address_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxCommitterServer).Address(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tx_committer.TxCommitter/Address",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxCommitterServer).Address(ctx, req.(*AddressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TxCommitter_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tx_committer.TxCommitter",
	HandlerType: (*TxCommitterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CommitTx",
			Handler:    _TxCommitter_CommitTx_Handler,
		},
		{
			MethodName: "Address",
			Handler:    _TxCommitter_Address_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tx_committer.proto",
}
