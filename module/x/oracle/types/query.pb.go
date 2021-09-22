// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: oracle/v1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type QueryCurrentEpochRequest struct {
}

func (m *QueryCurrentEpochRequest) Reset()         { *m = QueryCurrentEpochRequest{} }
func (m *QueryCurrentEpochRequest) String() string { return proto.CompactTextString(m) }
func (*QueryCurrentEpochRequest) ProtoMessage()    {}
func (*QueryCurrentEpochRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_34238c8dfdfcd7ec, []int{0}
}
func (m *QueryCurrentEpochRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCurrentEpochRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCurrentEpochRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCurrentEpochRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCurrentEpochRequest.Merge(m, src)
}
func (m *QueryCurrentEpochRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryCurrentEpochRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCurrentEpochRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCurrentEpochRequest proto.InternalMessageInfo

type QueryCurrentEpochResponse struct {
	Epoch *Epoch `protobuf:"bytes,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *QueryCurrentEpochResponse) Reset()         { *m = QueryCurrentEpochResponse{} }
func (m *QueryCurrentEpochResponse) String() string { return proto.CompactTextString(m) }
func (*QueryCurrentEpochResponse) ProtoMessage()    {}
func (*QueryCurrentEpochResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_34238c8dfdfcd7ec, []int{1}
}
func (m *QueryCurrentEpochResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCurrentEpochResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCurrentEpochResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCurrentEpochResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCurrentEpochResponse.Merge(m, src)
}
func (m *QueryCurrentEpochResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryCurrentEpochResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCurrentEpochResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCurrentEpochResponse proto.InternalMessageInfo

func (m *QueryCurrentEpochResponse) GetEpoch() *Epoch {
	if m != nil {
		return m.Epoch
	}
	return nil
}

type QueryEthFeeRequest struct {
}

func (m *QueryEthFeeRequest) Reset()         { *m = QueryEthFeeRequest{} }
func (m *QueryEthFeeRequest) String() string { return proto.CompactTextString(m) }
func (*QueryEthFeeRequest) ProtoMessage()    {}
func (*QueryEthFeeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_34238c8dfdfcd7ec, []int{2}
}
func (m *QueryEthFeeRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryEthFeeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryEthFeeRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryEthFeeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryEthFeeRequest.Merge(m, src)
}
func (m *QueryEthFeeRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryEthFeeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryEthFeeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryEthFeeRequest proto.InternalMessageInfo

type QueryEthFeeResponse struct {
	Min  github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,1,opt,name=min,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"min"`
	Fast github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=fast,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"fast"`
}

func (m *QueryEthFeeResponse) Reset()         { *m = QueryEthFeeResponse{} }
func (m *QueryEthFeeResponse) String() string { return proto.CompactTextString(m) }
func (*QueryEthFeeResponse) ProtoMessage()    {}
func (*QueryEthFeeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_34238c8dfdfcd7ec, []int{3}
}
func (m *QueryEthFeeResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryEthFeeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryEthFeeResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryEthFeeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryEthFeeResponse.Merge(m, src)
}
func (m *QueryEthFeeResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryEthFeeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryEthFeeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryEthFeeResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*QueryCurrentEpochRequest)(nil), "oracle.v1.QueryCurrentEpochRequest")
	proto.RegisterType((*QueryCurrentEpochResponse)(nil), "oracle.v1.QueryCurrentEpochResponse")
	proto.RegisterType((*QueryEthFeeRequest)(nil), "oracle.v1.QueryEthFeeRequest")
	proto.RegisterType((*QueryEthFeeResponse)(nil), "oracle.v1.QueryEthFeeResponse")
}

func init() { proto.RegisterFile("oracle/v1/query.proto", fileDescriptor_34238c8dfdfcd7ec) }

var fileDescriptor_34238c8dfdfcd7ec = []byte{
	// 398 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0xc1, 0xca, 0xd3, 0x40,
	0x18, 0x4c, 0xaa, 0x2d, 0x74, 0xf5, 0x20, 0x6b, 0x95, 0x18, 0xea, 0x56, 0x62, 0x29, 0x5e, 0xcc,
	0xd2, 0xfa, 0x02, 0xd2, 0x52, 0x41, 0xc4, 0x83, 0xc5, 0x93, 0x17, 0x49, 0xd2, 0xaf, 0x49, 0xb0,
	0xd9, 0x4d, 0xb3, 0x9b, 0x62, 0x4f, 0x82, 0x4f, 0x20, 0x78, 0xf4, 0x85, 0x7a, 0x2c, 0x78, 0x11,
	0x0f, 0x45, 0x5a, 0x5f, 0xc0, 0x37, 0x90, 0xec, 0xa6, 0x25, 0x2d, 0x7f, 0xff, 0xc3, 0x7f, 0xca,
	0x32, 0xdf, 0x37, 0x33, 0x3b, 0x93, 0x45, 0x0f, 0x78, 0xe6, 0x05, 0x73, 0xa0, 0xcb, 0x3e, 0x5d,
	0xe4, 0x90, 0xad, 0xdc, 0x34, 0xe3, 0x92, 0xe3, 0xa6, 0x86, 0xdd, 0x65, 0xdf, 0x6e, 0x87, 0x9c,
	0x87, 0x73, 0xa0, 0x5e, 0x1a, 0x53, 0x8f, 0x31, 0x2e, 0x3d, 0x19, 0x73, 0x26, 0xf4, 0xa2, 0x5d,
	0xe1, 0xcb, 0x55, 0x0a, 0x07, 0xb8, 0x15, 0xf2, 0x90, 0xab, 0x23, 0x2d, 0x4e, 0x1a, 0x75, 0x6c,
	0x64, 0xbd, 0x2b, 0x4c, 0x46, 0x79, 0x96, 0x01, 0x93, 0xe3, 0x94, 0x07, 0xd1, 0x04, 0x16, 0x39,
	0x08, 0xe9, 0x8c, 0xd0, 0xa3, 0x2b, 0x66, 0x22, 0xe5, 0x4c, 0x00, 0xee, 0xa1, 0x3a, 0x14, 0x80,
	0x65, 0x3e, 0x31, 0x9f, 0xdd, 0x19, 0xdc, 0x73, 0x8f, 0xd7, 0x73, 0xf5, 0xa2, 0x1e, 0x3b, 0x2d,
	0x84, 0x95, 0xc8, 0x58, 0x46, 0xaf, 0x00, 0x0e, 0xd2, 0x3f, 0x4c, 0x74, 0xff, 0x04, 0x2e, 0x55,
	0x5f, 0xa2, 0x5b, 0x49, 0xcc, 0x94, 0x66, 0x73, 0xe8, 0xae, 0xb7, 0x1d, 0xe3, 0xf7, 0xb6, 0xd3,
	0x0b, 0x63, 0x19, 0xe5, 0xbe, 0x1b, 0xf0, 0x84, 0x06, 0x5c, 0x24, 0x5c, 0x94, 0x9f, 0xe7, 0x62,
	0xfa, 0xa9, 0xcc, 0xf8, 0x9a, 0xc9, 0x49, 0x41, 0xc5, 0x43, 0x74, 0x7b, 0xe6, 0x09, 0x69, 0xd5,
	0x6e, 0x24, 0xa1, 0xb8, 0x83, 0x7f, 0x26, 0xaa, 0xab, 0xdb, 0xe1, 0x2f, 0xe8, 0x6e, 0x35, 0x3d,
	0x7e, 0x5a, 0x89, 0x79, 0xa9, 0x37, 0xbb, 0x7b, 0xfd, 0x92, 0x8e, 0xea, 0x74, 0xbf, 0xfe, 0xfc,
	0xfb, 0xbd, 0x46, 0x70, 0x9b, 0x1e, 0xff, 0x97, 0x0f, 0xd2, 0xa3, 0xaa, 0x36, 0x1a, 0x68, 0x0a,
	0x0e, 0x51, 0x43, 0x57, 0x84, 0x1f, 0x9f, 0xab, 0x9e, 0x34, 0x6a, 0x93, 0x4b, 0xe3, 0xd2, 0x8e,
	0x28, 0x3b, 0x0b, 0x3f, 0x3c, 0xb7, 0x93, 0xd1, 0xc7, 0x19, 0xc0, 0xf0, 0xcd, 0x7a, 0x47, 0xcc,
	0xcd, 0x8e, 0x98, 0x7f, 0x76, 0xc4, 0xfc, 0xb6, 0x27, 0xc6, 0x66, 0x4f, 0x8c, 0x5f, 0x7b, 0x62,
	0x7c, 0xe8, 0x57, 0xba, 0x7b, 0x1b, 0x33, 0x09, 0xd9, 0x7b, 0xf0, 0x12, 0x9a, 0x44, 0xb9, 0x3f,
	0xa0, 0x09, 0x9f, 0xe6, 0x73, 0xa0, 0x9f, 0x0f, 0xaa, 0xaa, 0x4a, 0xbf, 0xa1, 0x1e, 0xd7, 0x8b,
	0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0x07, 0x95, 0xc5, 0x23, 0xcb, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	CurrentEpoch(ctx context.Context, in *QueryCurrentEpochRequest, opts ...grpc.CallOption) (*QueryCurrentEpochResponse, error)
	EthFee(ctx context.Context, in *QueryEthFeeRequest, opts ...grpc.CallOption) (*QueryEthFeeResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) CurrentEpoch(ctx context.Context, in *QueryCurrentEpochRequest, opts ...grpc.CallOption) (*QueryCurrentEpochResponse, error) {
	out := new(QueryCurrentEpochResponse)
	err := c.cc.Invoke(ctx, "/oracle.v1.Query/CurrentEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) EthFee(ctx context.Context, in *QueryEthFeeRequest, opts ...grpc.CallOption) (*QueryEthFeeResponse, error) {
	out := new(QueryEthFeeResponse)
	err := c.cc.Invoke(ctx, "/oracle.v1.Query/EthFee", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	CurrentEpoch(context.Context, *QueryCurrentEpochRequest) (*QueryCurrentEpochResponse, error)
	EthFee(context.Context, *QueryEthFeeRequest) (*QueryEthFeeResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) CurrentEpoch(ctx context.Context, req *QueryCurrentEpochRequest) (*QueryCurrentEpochResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CurrentEpoch not implemented")
}
func (*UnimplementedQueryServer) EthFee(ctx context.Context, req *QueryEthFeeRequest) (*QueryEthFeeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EthFee not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_CurrentEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCurrentEpochRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).CurrentEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/oracle.v1.Query/CurrentEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).CurrentEpoch(ctx, req.(*QueryCurrentEpochRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_EthFee_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryEthFeeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).EthFee(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/oracle.v1.Query/EthFee",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).EthFee(ctx, req.(*QueryEthFeeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "oracle.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CurrentEpoch",
			Handler:    _Query_CurrentEpoch_Handler,
		},
		{
			MethodName: "EthFee",
			Handler:    _Query_EthFee_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "oracle/v1/query.proto",
}

func (m *QueryCurrentEpochRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCurrentEpochRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCurrentEpochRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryCurrentEpochResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCurrentEpochResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCurrentEpochResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Epoch != nil {
		{
			size, err := m.Epoch.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintQuery(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryEthFeeRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryEthFeeRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryEthFeeRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryEthFeeResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryEthFeeResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryEthFeeResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Fast.Size()
		i -= size
		if _, err := m.Fast.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size := m.Min.Size()
		i -= size
		if _, err := m.Min.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryCurrentEpochRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryCurrentEpochResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != nil {
		l = m.Epoch.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryEthFeeRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryEthFeeResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Min.Size()
	n += 1 + l + sovQuery(uint64(l))
	l = m.Fast.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryCurrentEpochRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryCurrentEpochRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCurrentEpochRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryCurrentEpochResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryCurrentEpochResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCurrentEpochResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Epoch == nil {
				m.Epoch = &Epoch{}
			}
			if err := m.Epoch.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryEthFeeRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryEthFeeRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryEthFeeRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryEthFeeResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryEthFeeResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryEthFeeResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Min", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Min.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fast", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Fast.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
