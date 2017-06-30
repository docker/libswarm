// Code generated by protoc-gen-gogo.
// source: github.com/containerd/containerd/api/services/diff/diff.proto
// DO NOT EDIT!

/*
	Package diff is a generated protocol buffer package.

	It is generated from these files:
		github.com/containerd/containerd/api/services/diff/diff.proto

	It has these top-level messages:
		ApplyRequest
		ApplyResponse
		DiffRequest
		DiffResponse
*/
package diff

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"
import _ "github.com/golang/protobuf/ptypes/empty"
import _ "github.com/gogo/protobuf/types"
import containerd_v1_types "github.com/containerd/containerd/api/types/mount"
import containerd_v1_types1 "github.com/containerd/containerd/api/types/descriptor"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type ApplyRequest struct {
	// Diff is the descriptor of the diff to be extracted
	Diff   *containerd_v1_types1.Descriptor `protobuf:"bytes,1,opt,name=diff" json:"diff,omitempty"`
	Mounts []*containerd_v1_types.Mount     `protobuf:"bytes,2,rep,name=mounts" json:"mounts,omitempty"`
}

func (m *ApplyRequest) Reset()                    { *m = ApplyRequest{} }
func (*ApplyRequest) ProtoMessage()               {}
func (*ApplyRequest) Descriptor() ([]byte, []int) { return fileDescriptorDiff, []int{0} }

type ApplyResponse struct {
	// Applied is the descriptor for the object which was applied.
	// If the input was a compressed blob then the result will be
	// the descriptor for the uncompressed blob.
	Applied *containerd_v1_types1.Descriptor `protobuf:"bytes,1,opt,name=applied" json:"applied,omitempty"`
}

func (m *ApplyResponse) Reset()                    { *m = ApplyResponse{} }
func (*ApplyResponse) ProtoMessage()               {}
func (*ApplyResponse) Descriptor() ([]byte, []int) { return fileDescriptorDiff, []int{1} }

type DiffRequest struct {
	// Left are the mounts which represent the older copy
	// in which is the base of the computed changes.
	Left []*containerd_v1_types.Mount `protobuf:"bytes,1,rep,name=left" json:"left,omitempty"`
	// Right are the mounts which represents the newer copy
	// in which changes from the left were made into.
	Right []*containerd_v1_types.Mount `protobuf:"bytes,2,rep,name=right" json:"right,omitempty"`
	// MediaType is the media type descriptor for the created diff
	// object
	MediaType string `protobuf:"bytes,3,opt,name=media_type,json=mediaType,proto3" json:"media_type,omitempty"`
	// Ref identifies the pre-commit content store object. This
	// reference can be used to get the status from the content store.
	Ref string `protobuf:"bytes,5,opt,name=ref,proto3" json:"ref,omitempty"`
}

func (m *DiffRequest) Reset()                    { *m = DiffRequest{} }
func (*DiffRequest) ProtoMessage()               {}
func (*DiffRequest) Descriptor() ([]byte, []int) { return fileDescriptorDiff, []int{2} }

type DiffResponse struct {
	// Diff is the descriptor of the diff which can be applied
	Diff *containerd_v1_types1.Descriptor `protobuf:"bytes,3,opt,name=diff" json:"diff,omitempty"`
}

func (m *DiffResponse) Reset()                    { *m = DiffResponse{} }
func (*DiffResponse) ProtoMessage()               {}
func (*DiffResponse) Descriptor() ([]byte, []int) { return fileDescriptorDiff, []int{3} }

func init() {
	proto.RegisterType((*ApplyRequest)(nil), "containerd.services.diff.v1.ApplyRequest")
	proto.RegisterType((*ApplyResponse)(nil), "containerd.services.diff.v1.ApplyResponse")
	proto.RegisterType((*DiffRequest)(nil), "containerd.services.diff.v1.DiffRequest")
	proto.RegisterType((*DiffResponse)(nil), "containerd.services.diff.v1.DiffResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Diff service

type DiffClient interface {
	// Apply applies the content associated with the provided digests onto
	// the provided mounts. Archive content will be extracted and
	// decompressed if necessary.
	Apply(ctx context.Context, in *ApplyRequest, opts ...grpc.CallOption) (*ApplyResponse, error)
	// Diff creates a diff between the given mounts and uploads the result
	// to the content store.
	Diff(ctx context.Context, in *DiffRequest, opts ...grpc.CallOption) (*DiffResponse, error)
}

type diffClient struct {
	cc *grpc.ClientConn
}

func NewDiffClient(cc *grpc.ClientConn) DiffClient {
	return &diffClient{cc}
}

func (c *diffClient) Apply(ctx context.Context, in *ApplyRequest, opts ...grpc.CallOption) (*ApplyResponse, error) {
	out := new(ApplyResponse)
	err := grpc.Invoke(ctx, "/containerd.services.diff.v1.Diff/Apply", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diffClient) Diff(ctx context.Context, in *DiffRequest, opts ...grpc.CallOption) (*DiffResponse, error) {
	out := new(DiffResponse)
	err := grpc.Invoke(ctx, "/containerd.services.diff.v1.Diff/Diff", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Diff service

type DiffServer interface {
	// Apply applies the content associated with the provided digests onto
	// the provided mounts. Archive content will be extracted and
	// decompressed if necessary.
	Apply(context.Context, *ApplyRequest) (*ApplyResponse, error)
	// Diff creates a diff between the given mounts and uploads the result
	// to the content store.
	Diff(context.Context, *DiffRequest) (*DiffResponse, error)
}

func RegisterDiffServer(s *grpc.Server, srv DiffServer) {
	s.RegisterService(&_Diff_serviceDesc, srv)
}

func _Diff_Apply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApplyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiffServer).Apply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/containerd.services.diff.v1.Diff/Apply",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiffServer).Apply(ctx, req.(*ApplyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Diff_Diff_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiffRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiffServer).Diff(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/containerd.services.diff.v1.Diff/Diff",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiffServer).Diff(ctx, req.(*DiffRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Diff_serviceDesc = grpc.ServiceDesc{
	ServiceName: "containerd.services.diff.v1.Diff",
	HandlerType: (*DiffServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Apply",
			Handler:    _Diff_Apply_Handler,
		},
		{
			MethodName: "Diff",
			Handler:    _Diff_Diff_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/containerd/containerd/api/services/diff/diff.proto",
}

func (m *ApplyRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ApplyRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Diff != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintDiff(dAtA, i, uint64(m.Diff.Size()))
		n1, err := m.Diff.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if len(m.Mounts) > 0 {
		for _, msg := range m.Mounts {
			dAtA[i] = 0x12
			i++
			i = encodeVarintDiff(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *ApplyResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ApplyResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Applied != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintDiff(dAtA, i, uint64(m.Applied.Size()))
		n2, err := m.Applied.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func (m *DiffRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DiffRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Left) > 0 {
		for _, msg := range m.Left {
			dAtA[i] = 0xa
			i++
			i = encodeVarintDiff(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Right) > 0 {
		for _, msg := range m.Right {
			dAtA[i] = 0x12
			i++
			i = encodeVarintDiff(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.MediaType) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintDiff(dAtA, i, uint64(len(m.MediaType)))
		i += copy(dAtA[i:], m.MediaType)
	}
	if len(m.Ref) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintDiff(dAtA, i, uint64(len(m.Ref)))
		i += copy(dAtA[i:], m.Ref)
	}
	return i, nil
}

func (m *DiffResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DiffResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Diff != nil {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintDiff(dAtA, i, uint64(m.Diff.Size()))
		n3, err := m.Diff.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n3
	}
	return i, nil
}

func encodeFixed64Diff(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Diff(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintDiff(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *ApplyRequest) Size() (n int) {
	var l int
	_ = l
	if m.Diff != nil {
		l = m.Diff.Size()
		n += 1 + l + sovDiff(uint64(l))
	}
	if len(m.Mounts) > 0 {
		for _, e := range m.Mounts {
			l = e.Size()
			n += 1 + l + sovDiff(uint64(l))
		}
	}
	return n
}

func (m *ApplyResponse) Size() (n int) {
	var l int
	_ = l
	if m.Applied != nil {
		l = m.Applied.Size()
		n += 1 + l + sovDiff(uint64(l))
	}
	return n
}

func (m *DiffRequest) Size() (n int) {
	var l int
	_ = l
	if len(m.Left) > 0 {
		for _, e := range m.Left {
			l = e.Size()
			n += 1 + l + sovDiff(uint64(l))
		}
	}
	if len(m.Right) > 0 {
		for _, e := range m.Right {
			l = e.Size()
			n += 1 + l + sovDiff(uint64(l))
		}
	}
	l = len(m.MediaType)
	if l > 0 {
		n += 1 + l + sovDiff(uint64(l))
	}
	l = len(m.Ref)
	if l > 0 {
		n += 1 + l + sovDiff(uint64(l))
	}
	return n
}

func (m *DiffResponse) Size() (n int) {
	var l int
	_ = l
	if m.Diff != nil {
		l = m.Diff.Size()
		n += 1 + l + sovDiff(uint64(l))
	}
	return n
}

func sovDiff(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozDiff(x uint64) (n int) {
	return sovDiff(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *ApplyRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ApplyRequest{`,
		`Diff:` + strings.Replace(fmt.Sprintf("%v", this.Diff), "Descriptor", "containerd_v1_types1.Descriptor", 1) + `,`,
		`Mounts:` + strings.Replace(fmt.Sprintf("%v", this.Mounts), "Mount", "containerd_v1_types.Mount", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *ApplyResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ApplyResponse{`,
		`Applied:` + strings.Replace(fmt.Sprintf("%v", this.Applied), "Descriptor", "containerd_v1_types1.Descriptor", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *DiffRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&DiffRequest{`,
		`Left:` + strings.Replace(fmt.Sprintf("%v", this.Left), "Mount", "containerd_v1_types.Mount", 1) + `,`,
		`Right:` + strings.Replace(fmt.Sprintf("%v", this.Right), "Mount", "containerd_v1_types.Mount", 1) + `,`,
		`MediaType:` + fmt.Sprintf("%v", this.MediaType) + `,`,
		`Ref:` + fmt.Sprintf("%v", this.Ref) + `,`,
		`}`,
	}, "")
	return s
}
func (this *DiffResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&DiffResponse{`,
		`Diff:` + strings.Replace(fmt.Sprintf("%v", this.Diff), "Descriptor", "containerd_v1_types1.Descriptor", 1) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringDiff(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *ApplyRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDiff
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ApplyRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ApplyRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Diff", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDiff
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Diff == nil {
				m.Diff = &containerd_v1_types1.Descriptor{}
			}
			if err := m.Diff.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Mounts", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDiff
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Mounts = append(m.Mounts, &containerd_v1_types.Mount{})
			if err := m.Mounts[len(m.Mounts)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDiff(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDiff
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
func (m *ApplyResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDiff
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ApplyResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ApplyResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Applied", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDiff
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Applied == nil {
				m.Applied = &containerd_v1_types1.Descriptor{}
			}
			if err := m.Applied.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDiff(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDiff
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
func (m *DiffRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDiff
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DiffRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DiffRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Left", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDiff
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Left = append(m.Left, &containerd_v1_types.Mount{})
			if err := m.Left[len(m.Left)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Right", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDiff
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Right = append(m.Right, &containerd_v1_types.Mount{})
			if err := m.Right[len(m.Right)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MediaType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDiff
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MediaType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ref", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDiff
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Ref = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDiff(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDiff
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
func (m *DiffResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDiff
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DiffResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DiffResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Diff", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDiff
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Diff == nil {
				m.Diff = &containerd_v1_types1.Descriptor{}
			}
			if err := m.Diff.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDiff(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDiff
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
func skipDiff(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowDiff
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
					return 0, ErrIntOverflowDiff
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
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
					return 0, ErrIntOverflowDiff
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthDiff
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowDiff
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
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
				next, err := skipDiff(dAtA[start:])
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
	ErrInvalidLengthDiff = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowDiff   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/containerd/containerd/api/services/diff/diff.proto", fileDescriptorDiff)
}

var fileDescriptorDiff = []byte{
	// 420 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x31, 0x8f, 0x94, 0x40,
	0x14, 0xc7, 0x6f, 0x64, 0xf7, 0xcc, 0xcd, 0x9e, 0x89, 0x99, 0x58, 0x10, 0x2e, 0x72, 0x1b, 0x2a,
	0xce, 0x62, 0xf0, 0xb8, 0xca, 0x44, 0x0b, 0xf5, 0x62, 0x61, 0x62, 0x43, 0xec, 0x4c, 0x34, 0x2c,
	0x3c, 0xb8, 0x49, 0x80, 0x19, 0x99, 0x61, 0x0d, 0x9d, 0x1f, 0xc5, 0xef, 0x62, 0x73, 0xa5, 0xa5,
	0xa5, 0xcb, 0x27, 0x31, 0x0c, 0x83, 0x12, 0x63, 0xf6, 0xd8, 0x66, 0xf2, 0x98, 0xf7, 0x7b, 0xef,
	0xfd, 0xdf, 0x9f, 0xc1, 0x2f, 0x72, 0xa6, 0x6e, 0x9a, 0x0d, 0x4d, 0x78, 0x19, 0x24, 0xbc, 0x52,
	0x31, 0xab, 0xa0, 0x4e, 0xa7, 0x61, 0x2c, 0x58, 0x20, 0xa1, 0xde, 0xb2, 0x04, 0x64, 0x90, 0xb2,
	0x2c, 0xd3, 0x07, 0x15, 0x35, 0x57, 0x9c, 0x9c, 0xfd, 0x05, 0xe9, 0x08, 0x51, 0x9d, 0xdf, 0x5e,
	0x3a, 0x8f, 0x72, 0x9e, 0x73, 0xcd, 0x05, 0x7d, 0x34, 0x94, 0x38, 0x67, 0x39, 0xe7, 0x79, 0x01,
	0x81, 0xfe, 0xda, 0x34, 0x59, 0x00, 0xa5, 0x50, 0xad, 0x49, 0x9e, 0xff, 0x9b, 0x54, 0xac, 0x04,
	0xa9, 0xe2, 0x52, 0x18, 0xe0, 0xf9, 0x2c, 0xbd, 0xaa, 0x15, 0x20, 0x83, 0x92, 0x37, 0x95, 0x1a,
	0x4e, 0x53, 0xfd, 0xe6, 0x80, 0xea, 0x14, 0x64, 0x52, 0x33, 0xa1, 0x78, 0x3d, 0x09, 0x87, 0x3e,
	0xde, 0x17, 0x7c, 0xfa, 0x52, 0x88, 0xa2, 0x8d, 0xe0, 0x73, 0x03, 0x52, 0x91, 0x2b, 0xbc, 0xe8,
	0x97, 0xb6, 0xd1, 0x1a, 0xf9, 0xab, 0xf0, 0x9c, 0x4e, 0x5c, 0xd9, 0x5e, 0x52, 0xdd, 0x8f, 0x5e,
	0xff, 0x69, 0x12, 0x69, 0x98, 0x84, 0xf8, 0x58, 0x6b, 0x93, 0xf6, 0xbd, 0xb5, 0xe5, 0xaf, 0x42,
	0xe7, 0xbf, 0x65, 0xef, 0x7a, 0x24, 0x32, 0xa4, 0xf7, 0x16, 0x3f, 0x30, 0x83, 0xa5, 0xe0, 0x95,
	0x04, 0xf2, 0x0c, 0xdf, 0x8f, 0x85, 0x28, 0x18, 0xa4, 0x73, 0x87, 0x8f, 0xbc, 0xf7, 0x0d, 0xe1,
	0xd5, 0x35, 0xcb, 0xb2, 0x71, 0x09, 0x8a, 0x17, 0x05, 0x64, 0xca, 0x46, 0x77, 0xaa, 0xd1, 0x1c,
	0x79, 0x8a, 0x97, 0x35, 0xcb, 0x6f, 0xd4, 0x0c, 0xf9, 0x03, 0x48, 0x1e, 0x63, 0x5c, 0x42, 0xca,
	0xe2, 0x4f, 0x7d, 0xce, 0xb6, 0xd6, 0xc8, 0x3f, 0x89, 0x4e, 0xf4, 0xcd, 0xfb, 0x56, 0x00, 0x79,
	0x88, 0xad, 0x1a, 0x32, 0x7b, 0xa9, 0xef, 0xfb, 0xd0, 0x7b, 0x8d, 0x4f, 0x07, 0x85, 0x66, 0xdb,
	0xd1, 0x67, 0xeb, 0x00, 0x9f, 0xc3, 0xef, 0x08, 0x2f, 0xfa, 0x2e, 0xe4, 0x23, 0x5e, 0x6a, 0xf3,
	0xc8, 0x05, 0xdd, 0xf3, 0x6c, 0xe9, 0xf4, 0xcf, 0x3a, 0x4f, 0xe6, 0xa0, 0x46, 0xdd, 0x07, 0x33,
	0xc7, 0xdf, 0x5b, 0x33, 0xb1, 0xdc, 0xb9, 0x98, 0x41, 0x0e, 0xcd, 0x5f, 0xd9, 0xb7, 0x3b, 0xf7,
	0xe8, 0xe7, 0xce, 0x3d, 0xfa, 0xda, 0xb9, 0xe8, 0xb6, 0x73, 0xd1, 0x8f, 0xce, 0x45, 0xbf, 0x3a,
	0x17, 0x6d, 0x8e, 0xf5, 0x9b, 0xbc, 0xfa, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x7f, 0xee, 0xae, 0xcf,
	0xcb, 0x03, 0x00, 0x00,
}