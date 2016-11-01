// Code generated by protoc-gen-go.
// source: online/mesg_online.proto
// DO NOT EDIT!

/*
Package mesg_online is a generated protocol buffer package.

It is generated from these files:
	online/mesg_online.proto

It has these top-level messages:
	MesgOnlineReq
	MesgOnlineAck
*/
package mesg_online

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type MesgOnlineReq struct {
	Uid              *uint64 `protobuf:"varint,1,opt,name=Uid" json:"Uid,omitempty"`
	Token            *string `protobuf:"bytes,2,opt,name=Token" json:"Token,omitempty"`
	App              *string `protobuf:"bytes,3,opt,name=App" json:"App,omitempty"`
	Version          *string `protobuf:"bytes,4,opt,name=Version" json:"Version,omitempty"`
	Terminal         *uint32 `protobuf:"varint,5,opt,name=Terminal" json:"Terminal,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *MesgOnlineReq) Reset()         { *m = MesgOnlineReq{} }
func (m *MesgOnlineReq) String() string { return proto.CompactTextString(m) }
func (*MesgOnlineReq) ProtoMessage()    {}

func (m *MesgOnlineReq) GetUid() uint64 {
	if m != nil && m.Uid != nil {
		return *m.Uid
	}
	return 0
}

func (m *MesgOnlineReq) GetToken() string {
	if m != nil && m.Token != nil {
		return *m.Token
	}
	return ""
}

func (m *MesgOnlineReq) GetApp() string {
	if m != nil && m.App != nil {
		return *m.App
	}
	return ""
}

func (m *MesgOnlineReq) GetVersion() string {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return ""
}

func (m *MesgOnlineReq) GetTerminal() uint32 {
	if m != nil && m.Terminal != nil {
		return *m.Terminal
	}
	return 0
}

type MesgOnlineAck struct {
	Uid              *uint64 `protobuf:"varint,1,opt,name=Uid" json:"Uid,omitempty"`
	Cid              *uint64 `protobuf:"varint,2,opt,name=Cid" json:"Cid,omitempty"`
	App              *string `protobuf:"bytes,3,opt,name=App" json:"App,omitempty"`
	Version          *string `protobuf:"bytes,4,opt,name=Version" json:"Version,omitempty"`
	Terminal         *uint32 `protobuf:"varint,5,opt,name=Terminal" json:"Terminal,omitempty"`
	Errnum           *uint32 `protobuf:"varint,6,opt,name=Errnum" json:"Errnum,omitempty"`
	Errmsg           *string `protobuf:"bytes,7,opt,name=Errmsg" json:"Errmsg,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *MesgOnlineAck) Reset()         { *m = MesgOnlineAck{} }
func (m *MesgOnlineAck) String() string { return proto.CompactTextString(m) }
func (*MesgOnlineAck) ProtoMessage()    {}

func (m *MesgOnlineAck) GetUid() uint64 {
	if m != nil && m.Uid != nil {
		return *m.Uid
	}
	return 0
}

func (m *MesgOnlineAck) GetCid() uint64 {
	if m != nil && m.Cid != nil {
		return *m.Cid
	}
	return 0
}

func (m *MesgOnlineAck) GetApp() string {
	if m != nil && m.App != nil {
		return *m.App
	}
	return ""
}

func (m *MesgOnlineAck) GetVersion() string {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return ""
}

func (m *MesgOnlineAck) GetTerminal() uint32 {
	if m != nil && m.Terminal != nil {
		return *m.Terminal
	}
	return 0
}

func (m *MesgOnlineAck) GetErrnum() uint32 {
	if m != nil && m.Errnum != nil {
		return *m.Errnum
	}
	return 0
}

func (m *MesgOnlineAck) GetErrmsg() string {
	if m != nil && m.Errmsg != nil {
		return *m.Errmsg
	}
	return ""
}
