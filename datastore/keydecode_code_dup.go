package datastore

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
)

// This code is duplicated from https://github.com/googleapis/google-cloud-go/blob/master/datastore/key.go with the method renamed.
// decodeToNewKey decodes a key from the opaque representation returned by Encode.
func decodeToNewKey(encoded string) (*NewFormatKey, error) {
	// Re-add padding.
	if m := len(encoded) % 4; m != 0 {
		encoded += strings.Repeat("=", 4-m)
	}

	b, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	pKey := new(PBKey)
	if err := proto.Unmarshal(b, pKey); err != nil {
		return nil, err
	}
	return cgdProtoToKey(pKey)
}

// protoToKey decodes a protocol buffer representation of a key into an
// equivalent *Key object. If the key is invalid, protoToKey will return the
// invalid key along with ErrInvalidKey.
func cgdProtoToKey(p *PBKey) (*NewFormatKey, error) {
	var key *NewFormatKey
	var namespace string
	if partition := p.PartitionId; partition != nil {
		namespace = partition.NamespaceId
	}
	for _, el := range p.Path {
		key = &NewFormatKey{
			Namespace: namespace,
			Kind:      el.Kind,
			ID:        el.GetId(),
			Name:      el.GetName(),
			Parent:    key,
		}
	}
	if !key.valid() { // Also detects key == nil.
		return key, ErrInvalidKey
	}
	return key, nil
}

// valid returns whether the key is valid.
func (k *NewFormatKey) valid() bool {
	if k == nil {
		return false
	}
	for ; k != nil; k = k.Parent {
		if k.Kind == "" {
			return false
		}
		if k.Name != "" && k.ID != 0 {
			return false
		}
		if k.Parent != nil {
			if k.Parent.Incomplete() {
				return false
			}
			if k.Parent.Namespace != k.Namespace {
				return false
			}
		}
	}
	return true
}

// Incomplete reports whether the key does not refer to a stored entity.
func (k *NewFormatKey) Incomplete() bool {
	return k.Name == "" && k.ID == 0
}

// This code is duplicated from https://github.com/googleapis/google-cloud-go/blob/master/datastore/key.go with the method renamed.
// Key represents the datastore key for a stored entity.
type NewFormatKey struct {
	// Kind cannot be empty.
	Kind string
	// Either ID or Name must be zero for the Key to be valid.
	// If both are zero, the Key is incomplete.
	ID   int64
	Name string
	// Parent must either be a complete Key or nil.
	Parent *NewFormatKey

	// Namespace provides the ability to partition your data for multiple
	// tenants. In most cases, it is not necessary to specify a namespace.
	// See docs on datastore multitenancy for details:
	// https://cloud.google.com/datastore/docs/concepts/multitenancy
	Namespace string
}

// A partition ID identifies a grouping of entities. The grouping is always
// by project and namespace, however the namespace ID may be empty.
//
// A partition ID contains several dimensions:
// project ID and namespace ID.
//
// Partition dimensions:
//
// - May be `""`.
// - Must be valid UTF-8 bytes.
// - Must have values that match regex `[A-Za-z\d\.\-_]{1,100}`
// If the value of any dimension matches regex `__.*__`, the partition is
// reserved/read-only.
// A reserved/read-only partition ID is forbidden in certain documented
// contexts.
//
// Foreign partition IDs (in which the project ID does
// not match the context project ID ) are discouraged.
// Reads and writes of foreign partition IDs may fail if the project is not in
// an active state.
type PartitionId struct {
	// The ID of the project to which the entities belong.
	ProjectId string `protobuf:"bytes,2,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	// If not empty, the ID of the namespace to which the entities belong.
	NamespaceId          string   `protobuf:"bytes,4,opt,name=namespace_id,json=namespaceId,proto3" json:"namespace_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PartitionId) Reset()         { *m = PartitionId{} }
func (m *PartitionId) String() string { return proto.CompactTextString(m) }
func (*PartitionId) ProtoMessage()    {}
func (*PartitionId) Descriptor() ([]byte, []int) {
	return fileDescriptor_entity_096a297364b049a5, []int{0}
}
func (m *PartitionId) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PartitionId.Unmarshal(m, b)
}
func (m *PartitionId) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PartitionId.Marshal(b, m, deterministic)
}
func (dst *PartitionId) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PartitionId.Merge(dst, src)
}
func (m *PartitionId) XXX_Size() int {
	return xxx_messageInfo_PartitionId.Size(m)
}
func (m *PartitionId) XXX_DiscardUnknown() {
	xxx_messageInfo_PartitionId.DiscardUnknown(m)
}

var xxx_messageInfo_PartitionId proto.InternalMessageInfo

func (m *PartitionId) GetProjectId() string {
	if m != nil {
		return m.ProjectId
	}
	return ""
}

func (m *PartitionId) GetNamespaceId() string {
	if m != nil {
		return m.NamespaceId
	}
	return ""
}

// A unique identifier for an entity.
// If a key's partition ID or any of its path kinds or names are
// reserved/read-only, the key is reserved/read-only.
// A reserved/read-only key is forbidden in certain documented contexts.
type PBKey struct {
	// Entities are partitioned into subsets, currently identified by a project
	// ID and namespace ID.
	// Queries are scoped to a single partition.
	PartitionId *PartitionId `protobuf:"bytes,1,opt,name=partition_id,json=partitionId,proto3" json:"partition_id,omitempty"`
	// The entity path.
	// An entity path consists of one or more elements composed of a kind and a
	// string or numerical identifier, which identify entities. The first
	// element identifies a _root entity_, the second element identifies
	// a _child_ of the root entity, the third element identifies a child of the
	// second entity, and so forth. The entities identified by all prefixes of
	// the path are called the element's _ancestors_.
	//
	// An entity path is always fully complete: *all* of the entity's ancestors
	// are required to be in the path along with the entity identifier itself.
	// The only exception is that in some documented cases, the identifier in the
	// last path element (for the entity) itself may be omitted. For example,
	// the last path element of the key of `Mutation.insert` may have no
	// identifier.
	//
	// A path can never be empty, and a path can have at most 100 elements.
	Path                 []*Key_PathElement `protobuf:"bytes,2,rep,name=path,proto3" json:"path,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *PBKey) Reset()         { *m = PBKey{} }
func (m *PBKey) String() string { return proto.CompactTextString(m) }
func (*PBKey) ProtoMessage()    {}
func (*PBKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_entity_096a297364b049a5, []int{1}
}
func (m *PBKey) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Key.Unmarshal(m, b)
}
func (m *PBKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Key.Marshal(b, m, deterministic)
}
func (dst *PBKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Key.Merge(dst, src)
}
func (m *PBKey) XXX_Size() int {
	return xxx_messageInfo_Key.Size(m)
}
func (m *PBKey) XXX_DiscardUnknown() {
	xxx_messageInfo_Key.DiscardUnknown(m)
}

// A (kind, ID/name) pair used to construct a key path.
//
// If either name or ID is set, the element is complete.
// If neither is set, the element is incomplete.
type Key_PathElement struct {
	// The kind of the entity.
	// A kind matching regex `__.*__` is reserved/read-only.
	// A kind must not contain more than 1500 bytes when UTF-8 encoded.
	// Cannot be `""`.
	Kind string `protobuf:"bytes,1,opt,name=kind,proto3" json:"kind,omitempty"`
	// The type of ID.
	//
	// Types that are valid to be assigned to IdType:
	//	*Key_PathElement_Id
	//	*Key_PathElement_Name
	IdType               isKey_PathElement_IdType `protobuf_oneof:"id_type"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *Key_PathElement) Reset()         { *m = Key_PathElement{} }
func (m *Key_PathElement) String() string { return proto.CompactTextString(m) }
func (*Key_PathElement) ProtoMessage()    {}
func (*Key_PathElement) Descriptor() ([]byte, []int) {
	return fileDescriptor_entity_096a297364b049a5, []int{1, 0}
}
func (m *Key_PathElement) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Key_PathElement.Unmarshal(m, b)
}
func (m *Key_PathElement) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Key_PathElement.Marshal(b, m, deterministic)
}
func (dst *Key_PathElement) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Key_PathElement.Merge(dst, src)
}
func (m *Key_PathElement) XXX_Size() int {
	return xxx_messageInfo_Key_PathElement.Size(m)
}
func (m *Key_PathElement) XXX_DiscardUnknown() {
	xxx_messageInfo_Key_PathElement.DiscardUnknown(m)
}

var xxx_messageInfo_Key_PathElement proto.InternalMessageInfo

func (m *Key_PathElement) GetKind() string {
	if m != nil {
		return m.Kind
	}
	return ""
}

type isKey_PathElement_IdType interface {
	isKey_PathElement_IdType()
}

type Key_PathElement_Id struct {
	Id int64 `protobuf:"varint,2,opt,name=id,proto3,oneof"`
}

type Key_PathElement_Name struct {
	Name string `protobuf:"bytes,3,opt,name=name,proto3,oneof"`
}

func (*Key_PathElement_Id) isKey_PathElement_IdType() {}

func (*Key_PathElement_Name) isKey_PathElement_IdType() {}

func (m *Key_PathElement) GetIdType() isKey_PathElement_IdType {
	if m != nil {
		return m.IdType
	}
	return nil
}

func (m *Key_PathElement) GetId() int64 {
	if x, ok := m.GetIdType().(*Key_PathElement_Id); ok {
		return x.Id
	}
	return 0
}

func (m *Key_PathElement) GetName() string {
	if x, ok := m.GetIdType().(*Key_PathElement_Name); ok {
		return x.Name
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Key_PathElement) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Key_PathElement_OneofMarshaler, _Key_PathElement_OneofUnmarshaler, _Key_PathElement_OneofSizer, []interface{}{
		(*Key_PathElement_Id)(nil),
		(*Key_PathElement_Name)(nil),
	}
}

func _Key_PathElement_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Key_PathElement)
	// id_type
	switch x := m.IdType.(type) {
	case *Key_PathElement_Id:
		b.EncodeVarint(2<<3 | proto.WireVarint)
		b.EncodeVarint(uint64(x.Id))
	case *Key_PathElement_Name:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.Name)
	case nil:
	default:
		return fmt.Errorf("Key_PathElement.IdType has unexpected type %T", x)
	}
	return nil
}

func _Key_PathElement_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Key_PathElement)
	switch tag {
	case 2: // id_type.id
		if wire != proto.WireVarint {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeVarint()
		m.IdType = &Key_PathElement_Id{int64(x)}
		return true, err
	case 3: // id_type.name
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.IdType = &Key_PathElement_Name{x}
		return true, err
	default:
		return false, nil
	}
}

func _Key_PathElement_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Key_PathElement)
	// id_type
	switch x := m.IdType.(type) {
	case *Key_PathElement_Id:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(x.Id))
	case *Key_PathElement_Name:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.Name)))
		n += len(x.Name)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

var fileDescriptor_entity_096a297364b049a5 = []byte{
	// 780 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x94, 0xff, 0x6e, 0xdc, 0x44,
	0x10, 0xc7, 0xed, 0xbb, 0x5c, 0x1a, 0x8f, 0xdd, 0xa4, 0x6c, 0x2a, 0x61, 0x02, 0x28, 0x26, 0x80,
	0x74, 0x02, 0xc9, 0x6e, 0xc2, 0x1f, 0x54, 0x14, 0xa4, 0x72, 0x25, 0xe0, 0x28, 0x15, 0x9c, 0x56,
	0x55, 0x24, 0x50, 0xa4, 0xd3, 0xde, 0x79, 0xeb, 0x2e, 0x67, 0xef, 0x5a, 0xf6, 0x3a, 0xaa, 0xdf,
	0x05, 0xf1, 0x00, 0x3c, 0x0a, 0x8f, 0x80, 0x78, 0x18, 0xb4, 0x3f, 0xec, 0x0b, 0xed, 0x35, 0xff,
	0x79, 0x67, 0x3e, 0xdf, 0xd9, 0xef, 0xec, 0xce, 0x1a, 0xa2, 0x5c, 0x88, 0xbc, 0xa0, 0x49, 0x46,
	0x24, 0x69, 0xa4, 0xa8, 0x69, 0x72, 0x73, 0x9a, 0x50, 0x2e, 0x99, 0xec, 0xe2, 0xaa, 0x16, 0x52,
	0xa0, 0x43, 0x43, 0xc4, 0x03, 0x11, 0xdf, 0x9c, 0x1e, 0x7d, 0x64, 0x65, 0xa4, 0x62, 0x09, 0xe1,
	0x5c, 0x48, 0x22, 0x99, 0xe0, 0x8d, 0x91, 0x0c, 0x59, 0xbd, 0x5a, 0xb6, 0x2f, 0x93, 0x46, 0xd6,
	0xed, 0x4a, 0xda, 0xec, 0xf1, 0x9b, 0x59, 0xc9, 0x4a, 0xda, 0x48, 0x52, 0x56, 0x16, 0x08, 0x2d,
	0x20, 0xbb, 0x8a, 0x26, 0x05, 0x91, 0x05, 0xcf, 0x4d, 0xe6, 0xe4, 0x17, 0xf0, 0xe7, 0xa4, 0x96,
	0x4c, 0x6d, 0x76, 0x91, 0xa1, 0x8f, 0x01, 0xaa, 0x5a, 0xfc, 0x4e, 0x57, 0x72, 0xc1, 0xb2, 0x70,
	0x14, 0xb9, 0x53, 0x0f, 0x7b, 0x36, 0x72, 0x91, 0xa1, 0x4f, 0x20, 0xe0, 0xa4, 0xa4, 0x4d, 0x45,
	0x56, 0x54, 0x01, 0x3b, 0x1a, 0xf0, 0x87, 0xd8, 0x45, 0x76, 0xf2, 0x8f, 0x0b, 0xe3, 0x4b, 0xda,
	0xa1, 0x67, 0x10, 0x54, 0x7d, 0x61, 0x85, 0xba, 0x91, 0x3b, 0xf5, 0xcf, 0xa2, 0x78, 0x4b, 0xef,
	0xf1, 0x2d, 0x07, 0xd8, 0xaf, 0x6e, 0xd9, 0x79, 0x0c, 0x3b, 0x15, 0x91, 0xaf, 0xc2, 0x51, 0x34,
	0x9e, 0xfa, 0x67, 0x9f, 0x6d, 0x15, 0x5f, 0xd2, 0x2e, 0x9e, 0x13, 0xf9, 0xea, 0xbc, 0xa0, 0x25,
	0xe5, 0x12, 0x6b, 0xc5, 0xd1, 0x0b, 0xd5, 0xd7, 0x10, 0x44, 0x08, 0x76, 0xd6, 0x8c, 0x1b, 0x17,
	0x1e, 0xd6, 0xdf, 0xe8, 0x01, 0x8c, 0x6c, 0x8f, 0xe3, 0xd4, 0xc1, 0x23, 0x96, 0xa1, 0x87, 0xb0,
	0xa3, 0x5a, 0x09, 0xc7, 0x8a, 0x4a, 0x1d, 0xac, 0x57, 0x33, 0x0f, 0xee, 0xb1, 0x6c, 0xa1, 0x8e,
	0xee, 0xe4, 0x29, 0xc0, 0xf7, 0x75, 0x4d, 0xba, 0x2b, 0x52, 0xb4, 0x14, 0x9d, 0xc1, 0xee, 0x8d,
	0xfa, 0x68, 0x42, 0x57, 0xfb, 0x3b, 0xda, 0xea, 0x4f, 0xb3, 0xd8, 0x92, 0x27, 0x7f, 0x4c, 0x60,
	0x62, 0xd4, 0x4f, 0x00, 0x78, 0x5b, 0x14, 0x0b, 0x9d, 0x08, 0xfd, 0xc8, 0x9d, 0xee, 0x6f, 0x2a,
	0xf4, 0x37, 0x19, 0xff, 0xdc, 0x16, 0x85, 0xe6, 0x53, 0x07, 0x7b, 0xbc, 0x5f, 0xa0, 0xcf, 0xe1,
	0xfe, 0x52, 0x88, 0x82, 0x12, 0x6e, 0xf5, 0xaa, 0xb1, 0xbd, 0xd4, 0xc1, 0x81, 0x0d, 0x0f, 0x18,
	0xe3, 0x92, 0xe6, 0xb4, 0xb6, 0x58, 0xdf, 0x6d, 0x60, 0xc3, 0x06, 0xfb, 0x14, 0x82, 0x4c, 0xb4,
	0xcb, 0x82, 0x5a, 0x4a, 0xf5, 0xef, 0xa6, 0x0e, 0xf6, 0x4d, 0xd4, 0x40, 0xe7, 0x70, 0x30, 0x8c,
	0x95, 0xe5, 0x40, 0xdf, 0xe9, 0xdb, 0xa6, 0x5f, 0xf4, 0x5c, 0xea, 0xe0, 0xfd, 0x41, 0x64, 0xca,
	0x7c, 0x0d, 0xde, 0x9a, 0x76, 0xb6, 0xc0, 0x44, 0x17, 0x08, 0xdf, 0x75, 0xaf, 0xa9, 0x83, 0xf7,
	0xd6, 0xb4, 0x1b, 0x4c, 0x36, 0xb2, 0x66, 0x3c, 0xb7, 0xda, 0xf7, 0xec, 0x25, 0xf9, 0x26, 0x6a,
	0xa0, 0x63, 0x80, 0x65, 0x21, 0x96, 0x16, 0x41, 0x91, 0x3b, 0x0d, 0xd4, 0xc1, 0xa9, 0x98, 0x01,
	0xbe, 0x83, 0x83, 0x9c, 0x8a, 0x45, 0x25, 0x18, 0x97, 0x96, 0xda, 0xd3, 0x26, 0x0e, 0x7b, 0x13,
	0xea, 0xa2, 0xe3, 0xe7, 0x44, 0x3e, 0xe7, 0x79, 0xea, 0xe0, 0xfb, 0x39, 0x15, 0x73, 0x05, 0x1b,
	0xf9, 0x53, 0x08, 0xcc, 0x53, 0xb6, 0xda, 0x5d, 0xad, 0xfd, 0x70, 0x6b, 0x03, 0xe7, 0x1a, 0x54,
	0x0e, 0x8d, 0xc4, 0x54, 0x98, 0x81, 0x4f, 0xd4, 0x08, 0xd9, 0x02, 0x9e, 0x2e, 0x70, 0xbc, 0xb5,
	0xc0, 0x66, 0xd4, 0x52, 0x07, 0x03, 0xd9, 0x0c, 0x5e, 0x08, 0xf7, 0x4a, 0x4a, 0x38, 0xe3, 0x79,
	0xb8, 0x1f, 0xb9, 0xd3, 0x09, 0xee, 0x97, 0xe8, 0x11, 0x3c, 0xa4, 0xaf, 0x57, 0x45, 0x9b, 0xd1,
	0xc5, 0xcb, 0x5a, 0x94, 0x0b, 0xc6, 0x33, 0xfa, 0x9a, 0x36, 0xe1, 0xa1, 0x1a, 0x0f, 0x8c, 0x6c,
	0xee, 0xc7, 0x5a, 0x94, 0x17, 0x26, 0x33, 0x0b, 0x00, 0xb4, 0x13, 0x33, 0xe0, 0xff, 0xba, 0xb0,
	0x6b, 0x7c, 0xa3, 0x2f, 0x60, 0xbc, 0xa6, 0x9d, 0x7d, 0xb7, 0xef, 0xbc, 0x22, 0xac, 0x20, 0x74,
	0xa9, 0x7f, 0x1b, 0x15, 0xad, 0x25, 0xa3, 0x4d, 0x38, 0xd6, 0xaf, 0xe1, 0xcb, 0x3b, 0x0e, 0x25,
	0x9e, 0x0f, 0xf4, 0x39, 0x97, 0x75, 0x87, 0x6f, 0xc9, 0x8f, 0x7e, 0x85, 0x83, 0x37, 0xd2, 0xe8,
	0xc1, 0xc6, 0x8b, 0x67, 0x76, 0x7c, 0x04, 0x93, 0xcd, 0x44, 0xdf, 0xfd, 0xf4, 0x0c, 0xf8, 0xcd,
	0xe8, 0xb1, 0x3b, 0xfb, 0xd3, 0x85, 0xf7, 0x57, 0xa2, 0xdc, 0x06, 0xcf, 0x7c, 0x63, 0x6d, 0xae,
	0x86, 0x78, 0xee, 0xfe, 0xf6, 0xad, 0x65, 0x72, 0x51, 0x10, 0x9e, 0xc7, 0xa2, 0xce, 0x93, 0x9c,
	0x72, 0x3d, 0xe2, 0x89, 0x49, 0x91, 0x8a, 0x35, 0xff, 0xfb, 0xcb, 0x3f, 0x19, 0x16, 0x7f, 0x8d,
	0x3e, 0xf8, 0xc9, 0xc8, 0x9f, 0x15, 0xa2, 0xcd, 0xe2, 0x1f, 0x86, 0x8d, 0xae, 0x4e, 0xff, 0xee,
	0x73, 0xd7, 0x3a, 0x77, 0x3d, 0xe4, 0xae, 0xaf, 0x4e, 0x97, 0xbb, 0x7a, 0x83, 0xaf, 0xfe, 0x0b,
	0x00, 0x00, 0xff, 0xff, 0xf3, 0xdd, 0x11, 0x96, 0x45, 0x06, 0x00, 0x00,
}

var xxx_messageInfo_Key proto.InternalMessageInfo
