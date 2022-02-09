// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type CallRequest struct {
	Dns                  string   `protobuf:"bytes,1,opt,name=dns,proto3" json:"dns,omitempty"`
	Params               string   `protobuf:"bytes,2,opt,name=params,proto3" json:"params,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CallRequest) Reset()         { *m = CallRequest{} }
func (m *CallRequest) String() string { return proto.CompactTextString(m) }
func (*CallRequest) ProtoMessage()    {}
func (*CallRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{0}
}

func (m *CallRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CallRequest.Unmarshal(m, b)
}
func (m *CallRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CallRequest.Marshal(b, m, deterministic)
}
func (m *CallRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CallRequest.Merge(m, src)
}
func (m *CallRequest) XXX_Size() int {
	return xxx_messageInfo_CallRequest.Size(m)
}
func (m *CallRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CallRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CallRequest proto.InternalMessageInfo

func (m *CallRequest) GetDns() string {
	if m != nil {
		return m.Dns
	}
	return ""
}

func (m *CallRequest) GetParams() string {
	if m != nil {
		return m.Params
	}
	return ""
}

type CallReply struct {
	Code                 int64    `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Msg                  string   `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	Data                 string   `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CallReply) Reset()         { *m = CallReply{} }
func (m *CallReply) String() string { return proto.CompactTextString(m) }
func (*CallReply) ProtoMessage()    {}
func (*CallReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{1}
}

func (m *CallReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CallReply.Unmarshal(m, b)
}
func (m *CallReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CallReply.Marshal(b, m, deterministic)
}
func (m *CallReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CallReply.Merge(m, src)
}
func (m *CallReply) XXX_Size() int {
	return xxx_messageInfo_CallReply.Size(m)
}
func (m *CallReply) XXX_DiscardUnknown() {
	xxx_messageInfo_CallReply.DiscardUnknown(m)
}

var xxx_messageInfo_CallReply proto.InternalMessageInfo

func (m *CallReply) GetCode() int64 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *CallReply) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func (m *CallReply) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

func init() {
	proto.RegisterType((*CallRequest)(nil), "pb.CallRequest")
	proto.RegisterType((*CallReply)(nil), "pb.CallReply")
}

func init() { proto.RegisterFile("service.proto", fileDescriptor_a0b84a42fa06f626) }

var fileDescriptor_a0b84a42fa06f626 = []byte{
	// 175 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52, 0x32, 0xe7,
	0xe2, 0x76, 0x4e, 0xcc, 0xc9, 0x09, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x2e, 0x11, 0x12, 0xe0, 0x62,
	0x4e, 0xc9, 0x2b, 0x96, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0x31, 0x85, 0xc4, 0xb8, 0xd8,
	0x0a, 0x12, 0x8b, 0x12, 0x73, 0x8b, 0x25, 0x98, 0xc0, 0x82, 0x50, 0x9e, 0x92, 0x2b, 0x17, 0x27,
	0x44, 0x63, 0x41, 0x4e, 0xa5, 0x90, 0x10, 0x17, 0x4b, 0x72, 0x7e, 0x4a, 0x2a, 0x58, 0x1f, 0x73,
	0x10, 0x98, 0x0d, 0x32, 0x2a, 0xb7, 0x38, 0x1d, 0xaa, 0x0b, 0xc4, 0x04, 0xa9, 0x4a, 0x49, 0x2c,
	0x49, 0x94, 0x60, 0x06, 0x0b, 0x81, 0xd9, 0x46, 0x86, 0x5c, 0xec, 0xc1, 0x10, 0x47, 0x09, 0xa9,
	0x71, 0xb1, 0x80, 0x4c, 0x14, 0xe2, 0xd7, 0x2b, 0x48, 0xd2, 0x43, 0x72, 0x94, 0x14, 0x2f, 0x42,
	0xa0, 0x20, 0xa7, 0xd2, 0x89, 0x2d, 0x8a, 0x45, 0xcf, 0xba, 0x20, 0x29, 0x89, 0x0d, 0xec, 0x0b,
	0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x2c, 0x5d, 0x75, 0x80, 0xd6, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ServiceClient is the client API for Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ServiceClient interface {
	Call(ctx context.Context, in *CallRequest, opts ...grpc.CallOption) (*CallReply, error)
}

type serviceClient struct {
	cc *grpc.ClientConn
}

func NewServiceClient(cc *grpc.ClientConn) ServiceClient {
	return &serviceClient{cc}
}

func (c *serviceClient) Call(ctx context.Context, in *CallRequest, opts ...grpc.CallOption) (*CallReply, error) {
	out := new(CallReply)
	err := c.cc.Invoke(ctx, "/pb.Service/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServiceServer is the server API for Service service.
type ServiceServer interface {
	Call(context.Context, *CallRequest) (*CallReply, error)
}

// UnimplementedServiceServer can be embedded to have forward compatible implementations.
type UnimplementedServiceServer struct {
}

func (*UnimplementedServiceServer) Call(ctx context.Context, req *CallRequest) (*CallReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}

func RegisterServiceServer(s *grpc.Server, srv ServiceServer) {
	s.RegisterService(&_Service_serviceDesc, srv)
}

func _Service_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Service/Call",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).Call(ctx, req.(*CallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Service_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Service",
	HandlerType: (*ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler:    _Service_Call_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}
