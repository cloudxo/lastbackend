//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2019] Last.Backend LLC
// All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains
// the property of Last.Backend LLC and its suppliers,
// if any.  The intellectual and technical concepts contained
// herein are proprietary to Last.Backend LLC
// and its suppliers and may be covered by Russian Federation and Foreign Patents,
// patents in process, and are protected by trade secret or copyright law.
// Dissemination of this information or reproduction of this material
// is strictly forbidden unless prior written permission is obtained
// from Last.Backend LLC.
//

package types

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
)

type ProxyMessage struct {
	Type               string                       `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Source             string                       `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
	TimeNano           int64                        `protobuf:"varint,3,opt,name=time_nano,json=timeNano,proto3" json:"time_nano,omitempty"`
	Line               []byte                       `protobuf:"bytes,4,opt,name=line,proto3" json:"line,omitempty"`
	Partial            bool                         `protobuf:"varint,5,opt,name=partial,proto3" json:"partial,omitempty"`
	PartialLogMetadata *PartialProxyMessageMetadata `protobuf:"bytes,6,opt,name=partial_log_metadata,json=partialLogMetadata" json:"partial_log_metadata,omitempty"`
}

type PartialProxyMessageMetadata struct {
	Last    bool   `protobuf:"varint,1,opt,name=last,proto3" json:"last,omitempty"`
	Id      string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Ordinal int32  `protobuf:"varint,3,opt,name=ordinal,proto3" json:"ordinal,omitempty"`
}

type LogMessage struct {
	Data             string            `json:"message"`
	ContainerId      string            `json:"container_id"`
	ContainerName    string            `json:"container_name"`
	ContainerType    string            `json:"container_type"`
	Selflink         string            `json:"selflink"`
	ContainerCreated JsonTime          `json:"container_created"`
	Tag              string            `json:"tag"`
	Extra            map[string]string `json:"extra"`
	Host             string            `json:"host"`
	Timestamp        JsonTime          `json:"timestamp"`
}

func (m *ProxyMessage) Reset()                    { *m = ProxyMessage{} }
func (m *ProxyMessage) String() string            { return proto.CompactTextString(m) }
func (*ProxyMessage) ProtoMessage()               {}
func (*ProxyMessage) Descriptor() ([]byte, []int) { return fileDescriptorEntry, []int{0} }

type JsonTime struct {
	time.Time
}

func (t JsonTime) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("\"%s\"", t.Format(time.RFC3339Nano))
	return []byte(str), nil
}

var fileDescriptorEntry = []byte{
	// 237 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x74, 0x90, 0xbd, 0x4a, 0x04, 0x31,
	0x14, 0x85, 0xb9, 0xb3, 0x3f, 0xce, 0xdc, 0x5d, 0x2c, 0x82, 0x68, 0x40, 0x18, 0xc2, 0x56, 0xa9,
	0xb6, 0xd0, 0x37, 0x10, 0x6c, 0x44, 0x45, 0xd2, 0x58, 0x0e, 0x57, 0x27, 0x2c, 0x81, 0xd9, 0xdc,
	0x21, 0x13, 0x0b, 0x1f, 0xcd, 0x37, 0xb0, 0xf4, 0x11, 0x64, 0x9e, 0x44, 0x26, 0x4e, 0xec, 0xec,
	0xce, 0x39, 0x5f, 0x8a, 0x2f, 0x17, 0x37, 0xd6, 0xc7, 0xf0, 0xbe, 0xef, 0x03, 0x47, 0xde, 0x7d,
	0x00, 0x96, 0xf7, 0x7c, 0xb8, 0x9d, 0x26, 0x71, 0x8e, 0xeb, 0x81, 0xdf, 0xc2, 0xab, 0x95, 0xa0,
	0x40, 0x57, 0x66, 0x6e, 0xe2, 0x12, 0xab, 0xe8, 0x8e, 0xb6, 0xf1, 0xe4, 0x59, 0x16, 0x0a, 0xf4,
	0xc2, 0x94, 0xd3, 0xf0, 0x48, 0x9e, 0x85, 0xc0, 0x65, 0xe7, 0xbc, 0x95, 0x0b, 0x05, 0x7a, 0x6b,
	0x52, 0x16, 0x12, 0x4f, 0x7a, 0x0a, 0xd1, 0x51, 0x27, 0x97, 0x0a, 0x74, 0x69, 0x72, 0x15, 0x77,
	0x78, 0x36, 0xc7, 0xa6, 0xe3, 0x43, 0x73, 0xb4, 0x91, 0x5a, 0x8a, 0x24, 0x57, 0x0a, 0xf4, 0xe6,
	0x4a, 0xee, 0x9f, 0x7e, 0x61, 0x56, 0x7a, 0x98, 0xb9, 0x11, 0xfd, 0x1f, 0xc8, 0xdb, 0xee, 0x19,
	0x2f, 0xfe, 0x79, 0x9e, 0xa4, 0x68, 0x88, 0xe9, 0x1f, 0xa5, 0x49, 0x59, 0x9c, 0x62, 0xe1, 0xda,
	0xa4, 0x5f, 0x99, 0xc2, 0xb5, 0x93, 0x24, 0x87, 0xd6, 0x79, 0xea, 0x92, 0xfb, 0xca, 0xe4, 0x7a,
	0xb3, 0xfd, 0x1c, 0x6b, 0xf8, 0x1a, 0x6b, 0xf8, 0x1e, 0x6b, 0x78, 0x59, 0xa7, 0x4b, 0x5d, 0xff,
	0x04, 0x00, 0x00, 0xff, 0xff, 0x8f, 0xed, 0x9f, 0xb6, 0x38, 0x01, 0x00, 0x00,
}
