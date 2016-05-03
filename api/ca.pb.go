// Code generated by protoc-gen-gogo.
// source: ca.proto
// DO NOT EDIT!

package api

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import strings "strings"
import github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
import sort "sort"
import strconv "strconv"
import reflect "reflect"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import raftpicker "github.com/docker/swarm-v2/manager/raftpicker"
import codes "google.golang.org/grpc/codes"
import metadata "google.golang.org/grpc/metadata"
import transport "google.golang.org/grpc/transport"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type CertificateStatusRequest struct {
	Token string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
}

func (m *CertificateStatusRequest) Reset()                    { *m = CertificateStatusRequest{} }
func (*CertificateStatusRequest) ProtoMessage()               {}
func (*CertificateStatusRequest) Descriptor() ([]byte, []int) { return fileDescriptorCa, []int{0} }

type CertificateStatusResponse struct {
	Status      *IssuanceStatus `protobuf:"bytes,1,opt,name=status" json:"status,omitempty"`
	Certificate []byte          `protobuf:"bytes,2,opt,name=certificate,proto3" json:"certificate,omitempty"`
}

func (m *CertificateStatusResponse) Reset()                    { *m = CertificateStatusResponse{} }
func (*CertificateStatusResponse) ProtoMessage()               {}
func (*CertificateStatusResponse) Descriptor() ([]byte, []int) { return fileDescriptorCa, []int{1} }

type IssueCertificateRequest struct {
	Role string `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	CSR  []byte `protobuf:"bytes,2,opt,name=csr,proto3" json:"csr,omitempty"`
}

func (m *IssueCertificateRequest) Reset()                    { *m = IssueCertificateRequest{} }
func (*IssueCertificateRequest) ProtoMessage()               {}
func (*IssueCertificateRequest) Descriptor() ([]byte, []int) { return fileDescriptorCa, []int{2} }

type IssueCertificateResponse struct {
	Token string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
}

func (m *IssueCertificateResponse) Reset()                    { *m = IssueCertificateResponse{} }
func (*IssueCertificateResponse) ProtoMessage()               {}
func (*IssueCertificateResponse) Descriptor() ([]byte, []int) { return fileDescriptorCa, []int{3} }

type GetRootCACertificateRequest struct {
}

func (m *GetRootCACertificateRequest) Reset()                    { *m = GetRootCACertificateRequest{} }
func (*GetRootCACertificateRequest) ProtoMessage()               {}
func (*GetRootCACertificateRequest) Descriptor() ([]byte, []int) { return fileDescriptorCa, []int{4} }

type GetRootCACertificateResponse struct {
	Certificate []byte `protobuf:"bytes,1,opt,name=certificate,proto3" json:"certificate,omitempty"`
}

func (m *GetRootCACertificateResponse) Reset()                    { *m = GetRootCACertificateResponse{} }
func (*GetRootCACertificateResponse) ProtoMessage()               {}
func (*GetRootCACertificateResponse) Descriptor() ([]byte, []int) { return fileDescriptorCa, []int{5} }

func init() {
	proto.RegisterType((*CertificateStatusRequest)(nil), "docker.cluster.api.CertificateStatusRequest")
	proto.RegisterType((*CertificateStatusResponse)(nil), "docker.cluster.api.CertificateStatusResponse")
	proto.RegisterType((*IssueCertificateRequest)(nil), "docker.cluster.api.IssueCertificateRequest")
	proto.RegisterType((*IssueCertificateResponse)(nil), "docker.cluster.api.IssueCertificateResponse")
	proto.RegisterType((*GetRootCACertificateRequest)(nil), "docker.cluster.api.GetRootCACertificateRequest")
	proto.RegisterType((*GetRootCACertificateResponse)(nil), "docker.cluster.api.GetRootCACertificateResponse")
}

func (m *CertificateStatusRequest) Copy() *CertificateStatusRequest {
	if m == nil {
		return nil
	}

	o := &CertificateStatusRequest{
		Token: m.Token,
	}

	return o
}

func (m *CertificateStatusResponse) Copy() *CertificateStatusResponse {
	if m == nil {
		return nil
	}

	o := &CertificateStatusResponse{
		Status:      m.Status.Copy(),
		Certificate: m.Certificate,
	}

	return o
}

func (m *IssueCertificateRequest) Copy() *IssueCertificateRequest {
	if m == nil {
		return nil
	}

	o := &IssueCertificateRequest{
		Role: m.Role,
		CSR:  m.CSR,
	}

	return o
}

func (m *IssueCertificateResponse) Copy() *IssueCertificateResponse {
	if m == nil {
		return nil
	}

	o := &IssueCertificateResponse{
		Token: m.Token,
	}

	return o
}

func (m *GetRootCACertificateRequest) Copy() *GetRootCACertificateRequest {
	if m == nil {
		return nil
	}

	o := &GetRootCACertificateRequest{}

	return o
}

func (m *GetRootCACertificateResponse) Copy() *GetRootCACertificateResponse {
	if m == nil {
		return nil
	}

	o := &GetRootCACertificateResponse{
		Certificate: m.Certificate,
	}

	return o
}

func (this *CertificateStatusRequest) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&api.CertificateStatusRequest{")
	s = append(s, "Token: "+fmt.Sprintf("%#v", this.Token)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *CertificateStatusResponse) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&api.CertificateStatusResponse{")
	if this.Status != nil {
		s = append(s, "Status: "+fmt.Sprintf("%#v", this.Status)+",\n")
	}
	s = append(s, "Certificate: "+fmt.Sprintf("%#v", this.Certificate)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *IssueCertificateRequest) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&api.IssueCertificateRequest{")
	s = append(s, "Role: "+fmt.Sprintf("%#v", this.Role)+",\n")
	s = append(s, "CSR: "+fmt.Sprintf("%#v", this.CSR)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *IssueCertificateResponse) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&api.IssueCertificateResponse{")
	s = append(s, "Token: "+fmt.Sprintf("%#v", this.Token)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *GetRootCACertificateRequest) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 4)
	s = append(s, "&api.GetRootCACertificateRequest{")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *GetRootCACertificateResponse) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&api.GetRootCACertificateResponse{")
	s = append(s, "Certificate: "+fmt.Sprintf("%#v", this.Certificate)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringCa(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func extensionToGoStringCa(e map[int32]github_com_gogo_protobuf_proto.Extension) string {
	if e == nil {
		return "nil"
	}
	s := "map[int32]proto.Extension{"
	keys := make([]int, 0, len(e))
	for k := range e {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	ss := []string{}
	for _, k := range keys {
		ss = append(ss, strconv.Itoa(k)+": "+e[int32(k)].GoString())
	}
	s += strings.Join(ss, ",") + "}"
	return s
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion2

// Client API for CA service

type CAClient interface {
	IssueCertificate(ctx context.Context, in *IssueCertificateRequest, opts ...grpc.CallOption) (*IssueCertificateResponse, error)
	CertificateStatus(ctx context.Context, in *CertificateStatusRequest, opts ...grpc.CallOption) (*CertificateStatusResponse, error)
	GetRootCACertificate(ctx context.Context, in *GetRootCACertificateRequest, opts ...grpc.CallOption) (*GetRootCACertificateResponse, error)
}

type cAClient struct {
	cc *grpc.ClientConn
}

func NewCAClient(cc *grpc.ClientConn) CAClient {
	return &cAClient{cc}
}

func (c *cAClient) IssueCertificate(ctx context.Context, in *IssueCertificateRequest, opts ...grpc.CallOption) (*IssueCertificateResponse, error) {
	out := new(IssueCertificateResponse)
	err := grpc.Invoke(ctx, "/docker.cluster.api.CA/IssueCertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) CertificateStatus(ctx context.Context, in *CertificateStatusRequest, opts ...grpc.CallOption) (*CertificateStatusResponse, error) {
	out := new(CertificateStatusResponse)
	err := grpc.Invoke(ctx, "/docker.cluster.api.CA/CertificateStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) GetRootCACertificate(ctx context.Context, in *GetRootCACertificateRequest, opts ...grpc.CallOption) (*GetRootCACertificateResponse, error) {
	out := new(GetRootCACertificateResponse)
	err := grpc.Invoke(ctx, "/docker.cluster.api.CA/GetRootCACertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for CA service

type CAServer interface {
	IssueCertificate(context.Context, *IssueCertificateRequest) (*IssueCertificateResponse, error)
	CertificateStatus(context.Context, *CertificateStatusRequest) (*CertificateStatusResponse, error)
	GetRootCACertificate(context.Context, *GetRootCACertificateRequest) (*GetRootCACertificateResponse, error)
}

func RegisterCAServer(s *grpc.Server, srv CAServer) {
	s.RegisterService(&_CA_serviceDesc, srv)
}

func _CA_IssueCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssueCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).IssueCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/docker.cluster.api.CA/IssueCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).IssueCertificate(ctx, req.(*IssueCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

<<<<<<< cc47d6cf701fda325c99d19001dc7cdb9f88e541
func _CA_GetRootCACertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
=======
func _CA_CertificateStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(CertificateStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(CAServer).CertificateStatus(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _CA_GetRootCACertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
>>>>>>> Changing Cert issuance to be Async, adding raft store for certs
	in := new(GetRootCACertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).GetRootCACertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/docker.cluster.api.CA/GetRootCACertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).GetRootCACertificate(ctx, req.(*GetRootCACertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _CA_serviceDesc = grpc.ServiceDesc{
	ServiceName: "docker.cluster.api.CA",
	HandlerType: (*CAServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "IssueCertificate",
			Handler:    _CA_IssueCertificate_Handler,
		},
		{
			MethodName: "CertificateStatus",
			Handler:    _CA_CertificateStatus_Handler,
		},
		{
			MethodName: "GetRootCACertificate",
			Handler:    _CA_GetRootCACertificate_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}

func (m *CertificateStatusRequest) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *CertificateStatusRequest) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Token) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintCa(data, i, uint64(len(m.Token)))
		i += copy(data[i:], m.Token)
	}
	return i, nil
}

func (m *CertificateStatusResponse) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *CertificateStatusResponse) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Status != nil {
		data[i] = 0xa
		i++
		i = encodeVarintCa(data, i, uint64(m.Status.Size()))
		n1, err := m.Status.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.Certificate != nil {
		if len(m.Certificate) > 0 {
			data[i] = 0x12
			i++
			i = encodeVarintCa(data, i, uint64(len(m.Certificate)))
			i += copy(data[i:], m.Certificate)
		}
	}
	return i, nil
}

func (m *IssueCertificateRequest) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *IssueCertificateRequest) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Role) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintCa(data, i, uint64(len(m.Role)))
		i += copy(data[i:], m.Role)
	}
	if len(m.CSR) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintCa(data, i, uint64(len(m.CSR)))
		i += copy(data[i:], m.CSR)
	}
	return i, nil
}

func (m *IssueCertificateResponse) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *IssueCertificateResponse) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Token) > 0 {
		data[i] = 0xa
		i++
<<<<<<< cc47d6cf701fda325c99d19001dc7cdb9f88e541
		i = encodeVarintCa(data, i, uint64(m.Status.Size()))
		n1, err := m.Status.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if len(m.CertificateChain) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintCa(data, i, uint64(len(m.CertificateChain)))
		i += copy(data[i:], m.CertificateChain)
=======
		i = encodeVarintCa(data, i, uint64(len(m.Token)))
		i += copy(data[i:], m.Token)
>>>>>>> Changing Cert issuance to be Async, adding raft store for certs
	}
	return i, nil
}

func (m *GetRootCACertificateRequest) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *GetRootCACertificateRequest) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *GetRootCACertificateResponse) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *GetRootCACertificateResponse) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Certificate) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintCa(data, i, uint64(len(m.Certificate)))
		i += copy(data[i:], m.Certificate)
	}
	return i, nil
}

func encodeFixed64Ca(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Ca(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintCa(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}

type raftProxyCAServer struct {
	local        CAServer
	connSelector *raftpicker.ConnSelector
	cluster      raftpicker.RaftCluster
}

func NewRaftProxyCAServer(local CAServer, connSelector *raftpicker.ConnSelector, cluster raftpicker.RaftCluster) CAServer {
	return &raftProxyCAServer{
		local:        local,
		cluster:      cluster,
		connSelector: connSelector,
	}
}

func (p *raftProxyCAServer) IssueCertificate(ctx context.Context, r *IssueCertificateRequest) (*IssueCertificateResponse, error) {

	if p.cluster.IsLeader() {
		return p.local.IssueCertificate(ctx, r)
	}
	var addr string
	s, ok := transport.StreamFromContext(ctx)
	if ok {
		addr = s.ServerTransport().RemoteAddr().String()
	}
	md, ok := metadata.FromContext(ctx)
	if ok && len(md["redirect"]) != 0 {
		return nil, grpc.Errorf(codes.ResourceExhausted, "more than one redirect to leader from: %s", md["redirect"])
	}
	if !ok {
		md = metadata.New(map[string]string{})
	}
	md["redirect"] = append(md["redirect"], addr)
	ctx = metadata.NewContext(ctx, md)

	conn, err := p.connSelector.Conn()
	if err != nil {
		return nil, err
	}
	return NewCAClient(conn).IssueCertificate(ctx, r)
}

func (p *raftProxyCAServer) CertificateStatus(ctx context.Context, r *CertificateStatusRequest) (*CertificateStatusResponse, error) {

	if p.cluster.IsLeader() {
		return p.local.CertificateStatus(ctx, r)
	}
	var addr string
	s, ok := transport.StreamFromContext(ctx)
	if ok {
		addr = s.ServerTransport().RemoteAddr().String()
	}
	md, ok := metadata.FromContext(ctx)
	if ok && len(md["redirect"]) != 0 {
		return nil, grpc.Errorf(codes.ResourceExhausted, "more than one redirect to leader from: %s", md["redirect"])
	}
	if !ok {
		md = metadata.New(map[string]string{})
	}
	md["redirect"] = append(md["redirect"], addr)
	ctx = metadata.NewContext(ctx, md)

	conn, err := p.connSelector.Conn()
	if err != nil {
		return nil, err
	}
	return NewCAClient(conn).CertificateStatus(ctx, r)
}

func (p *raftProxyCAServer) GetRootCACertificate(ctx context.Context, r *GetRootCACertificateRequest) (*GetRootCACertificateResponse, error) {

	if p.cluster.IsLeader() {
		return p.local.GetRootCACertificate(ctx, r)
	}
	var addr string
	s, ok := transport.StreamFromContext(ctx)
	if ok {
		addr = s.ServerTransport().RemoteAddr().String()
	}
	md, ok := metadata.FromContext(ctx)
	if ok && len(md["redirect"]) != 0 {
		return nil, grpc.Errorf(codes.ResourceExhausted, "more than one redirect to leader from: %s", md["redirect"])
	}
	if !ok {
		md = metadata.New(map[string]string{})
	}
	md["redirect"] = append(md["redirect"], addr)
	ctx = metadata.NewContext(ctx, md)

	conn, err := p.connSelector.Conn()
	if err != nil {
		return nil, err
	}
	return NewCAClient(conn).GetRootCACertificate(ctx, r)
}

func (m *CertificateStatusRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.Token)
	if l > 0 {
		n += 1 + l + sovCa(uint64(l))
	}
	return n
}

func (m *CertificateStatusResponse) Size() (n int) {
	var l int
	_ = l
	if m.Status != nil {
		l = m.Status.Size()
		n += 1 + l + sovCa(uint64(l))
	}
	if m.Certificate != nil {
		l = len(m.Certificate)
		if l > 0 {
			n += 1 + l + sovCa(uint64(l))
		}
	}
	return n
}

func (m *IssueCertificateRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.Role)
	if l > 0 {
		n += 1 + l + sovCa(uint64(l))
	}
	l = len(m.CSR)
	if l > 0 {
		n += 1 + l + sovCa(uint64(l))
	}
	return n
}

func (m *IssueCertificateResponse) Size() (n int) {
	var l int
	_ = l
	l = len(m.Token)
	if l > 0 {
		n += 1 + l + sovCa(uint64(l))
	}
<<<<<<< cc47d6cf701fda325c99d19001dc7cdb9f88e541
	l = len(m.CertificateChain)
	if l > 0 {
		n += 1 + l + sovCa(uint64(l))
	}
=======
>>>>>>> Changing Cert issuance to be Async, adding raft store for certs
	return n
}

func (m *GetRootCACertificateRequest) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *GetRootCACertificateResponse) Size() (n int) {
	var l int
	_ = l
	l = len(m.Certificate)
	if l > 0 {
		n += 1 + l + sovCa(uint64(l))
	}
	return n
}

func sovCa(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozCa(x uint64) (n int) {
	return sovCa(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *CertificateStatusRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&CertificateStatusRequest{`,
		`Token:` + fmt.Sprintf("%v", this.Token) + `,`,
		`}`,
	}, "")
	return s
}
func (this *CertificateStatusResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&CertificateStatusResponse{`,
		`Status:` + strings.Replace(fmt.Sprintf("%v", this.Status), "IssuanceStatus", "IssuanceStatus", 1) + `,`,
		`Certificate:` + fmt.Sprintf("%v", this.Certificate) + `,`,
		`}`,
	}, "")
	return s
}
func (this *IssueCertificateRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&IssueCertificateRequest{`,
		`Role:` + fmt.Sprintf("%v", this.Role) + `,`,
		`CSR:` + fmt.Sprintf("%v", this.CSR) + `,`,
		`}`,
	}, "")
	return s
}
func (this *IssueCertificateResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&IssueCertificateResponse{`,
		`Token:` + fmt.Sprintf("%v", this.Token) + `,`,
		`}`,
	}, "")
	return s
}
func (this *GetRootCACertificateRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&GetRootCACertificateRequest{`,
		`}`,
	}, "")
	return s
}
func (this *GetRootCACertificateResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&GetRootCACertificateResponse{`,
		`Certificate:` + fmt.Sprintf("%v", this.Certificate) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringCa(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *CertificateStatusRequest) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCa
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CertificateStatusRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CertificateStatusRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCa
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCa
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Token = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCa(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCa
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
func (m *CertificateStatusResponse) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCa
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CertificateStatusResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CertificateStatusResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCa
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCa
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Status == nil {
				m.Status = &IssuanceStatus{}
			}
			if err := m.Status.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Certificate", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCa
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthCa
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Certificate = append(m.Certificate[:0], data[iNdEx:postIndex]...)
			if m.Certificate == nil {
				m.Certificate = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCa(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCa
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
func (m *IssueCertificateRequest) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCa
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: IssueCertificateRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IssueCertificateRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Role", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCa
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCa
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Role = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CSR", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCa
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthCa
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CSR = append(m.CSR[:0], data[iNdEx:postIndex]...)
			if m.CSR == nil {
				m.CSR = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCa(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCa
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
func (m *IssueCertificateResponse) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCa
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: IssueCertificateResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IssueCertificateResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCa
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCa
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Token = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCa(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCa
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
func (m *GetRootCACertificateRequest) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCa
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GetRootCACertificateRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetRootCACertificateRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipCa(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCa
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
func (m *GetRootCACertificateResponse) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCa
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GetRootCACertificateResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetRootCACertificateResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Certificate", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCa
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthCa
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Certificate = append(m.Certificate[:0], data[iNdEx:postIndex]...)
			if m.Certificate == nil {
				m.Certificate = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCa(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCa
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
func skipCa(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCa
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
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
					return 0, ErrIntOverflowCa
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCa
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthCa
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowCa
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipCa(data[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthCa = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCa   = fmt.Errorf("proto: integer overflow")
)

var fileDescriptorCa = []byte{
	// 351 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x53, 0x3d, 0x4f, 0x02, 0x41,
	0x10, 0xe5, 0x40, 0x51, 0x07, 0x0b, 0x9d, 0x90, 0x08, 0x27, 0x22, 0xb9, 0xca, 0x44, 0x3d, 0x08,
	0x76, 0x56, 0xc2, 0x15, 0x6a, 0xbb, 0xfc, 0x82, 0x73, 0x5d, 0x09, 0x81, 0xb0, 0xe7, 0xee, 0x5e,
	0x41, 0x6c, 0xfc, 0x79, 0x94, 0x96, 0x56, 0x46, 0x28, 0xac, 0xfd, 0x09, 0xee, 0x7d, 0xa0, 0x84,
	0x5b, 0x0c, 0xc5, 0xe4, 0xe6, 0xe6, 0xe6, 0xbd, 0x79, 0x6f, 0x6e, 0x17, 0x76, 0xa9, 0xef, 0x06,
	0x82, 0x2b, 0x8e, 0xf8, 0xc8, 0xe9, 0x90, 0x09, 0x97, 0x8e, 0x42, 0xa9, 0xf4, 0xd3, 0x0f, 0x06,
	0x76, 0x49, 0x4d, 0x02, 0x26, 0x93, 0x06, 0xbb, 0xdc, 0xe7, 0x7d, 0x1e, 0xa7, 0xcd, 0x28, 0x4b,
	0xaa, 0x4e, 0x0b, 0x2a, 0x1e, 0x13, 0x6a, 0xf0, 0x34, 0xa0, 0xbe, 0x62, 0x3d, 0xe5, 0xab, 0x50,
	0x12, 0xf6, 0x1c, 0x32, 0xa9, 0xb0, 0x0c, 0xdb, 0x8a, 0x0f, 0xd9, 0xb8, 0x62, 0x35, 0xac, 0xb3,
	0x3d, 0x92, 0xbc, 0x38, 0x13, 0xa8, 0x1a, 0x10, 0x32, 0xe0, 0x63, 0xc9, 0xf0, 0x1a, 0x8a, 0x32,
	0xae, 0xc4, 0x98, 0x52, 0xdb, 0x71, 0xb3, 0xb2, 0xdc, 0x7b, 0x29, 0x43, 0x7f, 0x4c, 0x17, 0xd8,
	0x14, 0x81, 0x0d, 0x28, 0xd1, 0x3f, 0xe2, 0x4a, 0x5e, 0x13, 0xec, 0x93, 0xe5, 0x92, 0x73, 0x07,
	0x47, 0x11, 0x96, 0x2d, 0xcd, 0x5f, 0x68, 0x45, 0xd8, 0x12, 0x7c, 0xc4, 0x52, 0xa9, 0x71, 0x8e,
	0x55, 0x28, 0x50, 0x29, 0x12, 0xa2, 0xee, 0xce, 0xfc, 0xe3, 0xb4, 0xe0, 0xf5, 0x08, 0x89, 0x6a,
	0x91, 0xed, 0x2c, 0x53, 0xea, 0xc1, 0x6c, 0xfb, 0x04, 0x8e, 0x6f, 0x99, 0x22, 0x9c, 0x2b, 0xaf,
	0x93, 0x9d, 0xef, 0xdc, 0x40, 0xcd, 0xfc, 0x39, 0x25, 0x5d, 0x31, 0x67, 0x65, 0xcc, 0xb5, 0xbf,
	0xf2, 0x90, 0xf7, 0x3a, 0xc8, 0xe1, 0x60, 0x55, 0x19, 0x9e, 0xaf, 0xdb, 0xa2, 0x61, 0x13, 0xf6,
	0xc5, 0x66, 0xcd, 0x89, 0x2e, 0x27, 0x87, 0x02, 0x0e, 0x33, 0xff, 0x13, 0x8d, 0x24, 0xeb, 0x0e,
	0x8a, 0x7d, 0xb9, 0x61, 0xf7, 0xef, 0xcc, 0x17, 0x28, 0x9b, 0xb6, 0x85, 0x4d, 0x13, 0xd1, 0x3f,
	0x6b, 0xb7, 0x5b, 0x9b, 0x03, 0x16, 0xc3, 0xbb, 0xb5, 0xe9, 0xac, 0x9e, 0x7b, 0xd7, 0xf1, 0x3d,
	0xab, 0x5b, 0xaf, 0xf3, 0xba, 0x35, 0xd5, 0xf1, 0xa6, 0xe3, 0x53, 0xc7, 0x43, 0x31, 0xbe, 0x17,
	0x57, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x3e, 0x4c, 0x96, 0x61, 0x5a, 0x03, 0x00, 0x00,
}
