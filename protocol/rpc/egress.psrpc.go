// Code generated by protoc-gen-psrpc v0.5.1, DO NOT EDIT.
// source: rpc/egress.proto

package rpc

import (
	"context"

	"github.com/livekit/psrpc"
	"github.com/livekit/psrpc/pkg/client"
	"github.com/livekit/psrpc/pkg/info"
	"github.com/livekit/psrpc/pkg/rand"
	"github.com/livekit/psrpc/pkg/server"
	"github.com/livekit/psrpc/version"
)
import livekit2 "jumat/protocol/livekit"

var _ = version.PsrpcVersion_0_5

// ===============================
// EgressInternal Client Interface
// ===============================

type EgressInternalClient interface {
	StartEgress(ctx context.Context, topic string, req *StartEgressRequest, opts ...psrpc.RequestOption) (*livekit2.EgressInfo, error)

	ListActiveEgress(ctx context.Context, req *ListActiveEgressRequest, opts ...psrpc.RequestOption) (<-chan *psrpc.Response[*ListActiveEgressResponse], error)
}

// ===================================
// EgressInternal ServerImpl Interface
// ===================================

type EgressInternalServerImpl interface {
	StartEgress(context.Context, *StartEgressRequest) (*livekit2.EgressInfo, error)
	StartEgressAffinity(context.Context, *StartEgressRequest) float32

	ListActiveEgress(context.Context, *ListActiveEgressRequest) (*ListActiveEgressResponse, error)
}

// ===============================
// EgressInternal Server Interface
// ===============================

type EgressInternalServer interface {
	RegisterStartEgressTopic(topic string) error
	DeregisterStartEgressTopic(topic string)

	// Close and wait for pending RPCs to complete
	Shutdown()

	// Close immediately, without waiting for pending RPCs
	Kill()
}

// =====================
// EgressInternal Client
// =====================

type egressInternalClient struct {
	client *client.RPCClient
}

// NewEgressInternalClient creates a psrpc client that implements the EgressInternalClient interface.
func NewEgressInternalClient(bus psrpc.MessageBus, opts ...psrpc.ClientOption) (EgressInternalClient, error) {
	sd := &info.ServiceDefinition{
		Name: "EgressInternal",
		ID:   rand.NewClientID(),
	}

	sd.RegisterMethod("StartEgress", true, false, true, false)
	sd.RegisterMethod("ListActiveEgress", false, true, false, false)

	rpcClient, err := client.NewRPCClient(sd, bus, opts...)
	if err != nil {
		return nil, err
	}

	return &egressInternalClient{
		client: rpcClient,
	}, nil
}

func (c *egressInternalClient) StartEgress(ctx context.Context, topic string, req *StartEgressRequest, opts ...psrpc.RequestOption) (*livekit2.EgressInfo, error) {
	return client.RequestSingle[*livekit2.EgressInfo](ctx, c.client, "StartEgress", []string{topic}, req, opts...)
}

func (c *egressInternalClient) ListActiveEgress(ctx context.Context, req *ListActiveEgressRequest, opts ...psrpc.RequestOption) (<-chan *psrpc.Response[*ListActiveEgressResponse], error) {
	return client.RequestMulti[*ListActiveEgressResponse](ctx, c.client, "ListActiveEgress", nil, req, opts...)
}

// =====================
// EgressInternal Server
// =====================

type egressInternalServer struct {
	svc EgressInternalServerImpl
	rpc *server.RPCServer
}

// NewEgressInternalServer builds a RPCServer that will route requests
// to the corresponding method in the provided svc implementation.
func NewEgressInternalServer(svc EgressInternalServerImpl, bus psrpc.MessageBus, opts ...psrpc.ServerOption) (EgressInternalServer, error) {
	sd := &info.ServiceDefinition{
		Name: "EgressInternal",
		ID:   rand.NewServerID(),
	}

	s := server.NewRPCServer(sd, bus, opts...)

	sd.RegisterMethod("StartEgress", true, false, true, false)
	sd.RegisterMethod("ListActiveEgress", false, true, false, false)
	var err error
	err = server.RegisterHandler(s, "ListActiveEgress", nil, svc.ListActiveEgress, nil)
	if err != nil {
		s.Close(false)
		return nil, err
	}

	return &egressInternalServer{
		svc: svc,
		rpc: s,
	}, nil
}

func (s *egressInternalServer) RegisterStartEgressTopic(topic string) error {
	return server.RegisterHandler(s.rpc, "StartEgress", []string{topic}, s.svc.StartEgress, s.svc.StartEgressAffinity)
}

func (s *egressInternalServer) DeregisterStartEgressTopic(topic string) {
	s.rpc.DeregisterHandler("StartEgress", []string{topic})
}

func (s *egressInternalServer) Shutdown() {
	s.rpc.Close(false)
}

func (s *egressInternalServer) Kill() {
	s.rpc.Close(true)
}

// ==============================
// EgressHandler Client Interface
// ==============================

type EgressHandlerClient interface {
	UpdateStream(ctx context.Context, topic string, req *livekit2.UpdateStreamRequest, opts ...psrpc.RequestOption) (*livekit2.EgressInfo, error)

	StopEgress(ctx context.Context, topic string, req *livekit2.StopEgressRequest, opts ...psrpc.RequestOption) (*livekit2.EgressInfo, error)
}

// ==================================
// EgressHandler ServerImpl Interface
// ==================================

type EgressHandlerServerImpl interface {
	UpdateStream(context.Context, *livekit2.UpdateStreamRequest) (*livekit2.EgressInfo, error)

	StopEgress(context.Context, *livekit2.StopEgressRequest) (*livekit2.EgressInfo, error)
}

// ==============================
// EgressHandler Server Interface
// ==============================

type EgressHandlerServer interface {
	RegisterUpdateStreamTopic(topic string) error
	DeregisterUpdateStreamTopic(topic string)
	RegisterStopEgressTopic(topic string) error
	DeregisterStopEgressTopic(topic string)

	// Close and wait for pending RPCs to complete
	Shutdown()

	// Close immediately, without waiting for pending RPCs
	Kill()
}

// ====================
// EgressHandler Client
// ====================

type egressHandlerClient struct {
	client *client.RPCClient
}

// NewEgressHandlerClient creates a psrpc client that implements the EgressHandlerClient interface.
func NewEgressHandlerClient(bus psrpc.MessageBus, opts ...psrpc.ClientOption) (EgressHandlerClient, error) {
	sd := &info.ServiceDefinition{
		Name: "EgressHandler",
		ID:   rand.NewClientID(),
	}

	sd.RegisterMethod("UpdateStream", false, false, true, true)
	sd.RegisterMethod("StopEgress", false, false, true, true)

	rpcClient, err := client.NewRPCClient(sd, bus, opts...)
	if err != nil {
		return nil, err
	}

	return &egressHandlerClient{
		client: rpcClient,
	}, nil
}

func (c *egressHandlerClient) UpdateStream(ctx context.Context, topic string, req *livekit2.UpdateStreamRequest, opts ...psrpc.RequestOption) (*livekit2.EgressInfo, error) {
	return client.RequestSingle[*livekit2.EgressInfo](ctx, c.client, "UpdateStream", []string{topic}, req, opts...)
}

func (c *egressHandlerClient) StopEgress(ctx context.Context, topic string, req *livekit2.StopEgressRequest, opts ...psrpc.RequestOption) (*livekit2.EgressInfo, error) {
	return client.RequestSingle[*livekit2.EgressInfo](ctx, c.client, "StopEgress", []string{topic}, req, opts...)
}

// ====================
// EgressHandler Server
// ====================

type egressHandlerServer struct {
	svc EgressHandlerServerImpl
	rpc *server.RPCServer
}

// NewEgressHandlerServer builds a RPCServer that will route requests
// to the corresponding method in the provided svc implementation.
func NewEgressHandlerServer(svc EgressHandlerServerImpl, bus psrpc.MessageBus, opts ...psrpc.ServerOption) (EgressHandlerServer, error) {
	sd := &info.ServiceDefinition{
		Name: "EgressHandler",
		ID:   rand.NewServerID(),
	}

	s := server.NewRPCServer(sd, bus, opts...)

	sd.RegisterMethod("UpdateStream", false, false, true, true)
	sd.RegisterMethod("StopEgress", false, false, true, true)
	return &egressHandlerServer{
		svc: svc,
		rpc: s,
	}, nil
}

func (s *egressHandlerServer) RegisterUpdateStreamTopic(topic string) error {
	return server.RegisterHandler(s.rpc, "UpdateStream", []string{topic}, s.svc.UpdateStream, nil)
}

func (s *egressHandlerServer) DeregisterUpdateStreamTopic(topic string) {
	s.rpc.DeregisterHandler("UpdateStream", []string{topic})
}

func (s *egressHandlerServer) RegisterStopEgressTopic(topic string) error {
	return server.RegisterHandler(s.rpc, "StopEgress", []string{topic}, s.svc.StopEgress, nil)
}

func (s *egressHandlerServer) DeregisterStopEgressTopic(topic string) {
	s.rpc.DeregisterHandler("StopEgress", []string{topic})
}

func (s *egressHandlerServer) Shutdown() {
	s.rpc.Close(false)
}

func (s *egressHandlerServer) Kill() {
	s.rpc.Close(true)
}

var psrpcFileDescriptor1 = []byte{
	// 587 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0x6f, 0x6f, 0x12, 0x31,
	0x18, 0xb7, 0x30, 0x18, 0x3c, 0x8c, 0x49, 0xea, 0xcc, 0xba, 0xdb, 0x96, 0x20, 0x6a, 0x42, 0x16,
	0x3d, 0x0c, 0x7b, 0xa3, 0xbe, 0xda, 0x66, 0x88, 0x23, 0x99, 0x51, 0x6f, 0x2e, 0x46, 0xdf, 0x90,
	0x72, 0xd7, 0x61, 0xc3, 0x71, 0xad, 0x6d, 0x61, 0xe1, 0x23, 0xec, 0x63, 0xf8, 0x15, 0x88, 0x9f,
	0xc8, 0x4f, 0x62, 0x68, 0xe1, 0x06, 0x2c, 0x18, 0x5f, 0xdd, 0xf5, 0xf7, 0xaf, 0xcf, 0xf3, 0xf4,
	0x81, 0x8a, 0x92, 0x61, 0x83, 0xf5, 0x14, 0xd3, 0xda, 0x97, 0x4a, 0x18, 0x81, 0xb3, 0x4a, 0x86,
	0xde, 0x5e, 0x4f, 0x88, 0x5e, 0xcc, 0x1a, 0x16, 0xea, 0x0e, 0xaf, 0x1b, 0x34, 0x19, 0x3b, 0xde,
	0x2b, 0x0b, 0x69, 0xb8, 0x48, 0x66, 0x72, 0x6f, 0x27, 0xe6, 0x23, 0xd6, 0xe7, 0xa6, 0xb3, 0x18,
	0x52, 0xfb, 0xb3, 0x01, 0xf8, 0xd2, 0x50, 0x65, 0x5a, 0x16, 0x0d, 0xd8, 0xcf, 0x21, 0xd3, 0x06,
	0xef, 0x43, 0xd1, 0xc9, 0x3a, 0x3c, 0x22, 0xa8, 0x8a, 0xea, 0xc5, 0xa0, 0xe0, 0x80, 0x76, 0x84,
	0x2f, 0x60, 0x5b, 0x09, 0x31, 0xe8, 0x84, 0x62, 0x20, 0x85, 0xe6, 0x86, 0x91, 0x5c, 0x15, 0xd5,
	0x4b, 0xcd, 0xa7, 0xfe, 0xec, 0x0a, 0x3f, 0x10, 0x62, 0xf0, 0x6e, 0xce, 0x2e, 0x25, 0x9f, 0x3f,
	0x08, 0xca, 0x6a, 0x91, 0xc5, 0x2f, 0x21, 0x7b, 0xc3, 0xba, 0xa4, 0x64, 0x23, 0xf6, 0xd2, 0x88,
	0xaf, 0xac, 0xbb, 0x6a, 0x9c, 0xea, 0x70, 0x0b, 0x4a, 0x92, 0x2a, 0xc3, 0x43, 0x2e, 0x69, 0x62,
	0x48, 0xd9, 0xda, 0x9e, 0xa4, 0xb6, 0x4f, 0x77, 0xdc, 0xaa, 0x7d, 0xd1, 0x87, 0x3f, 0xc2, 0x43,
	0xa3, 0x68, 0xd8, 0x5f, 0x68, 0x22, 0x6f, 0xa3, 0x9e, 0xa5, 0x51, 0x5f, 0xa6, 0xfc, 0xda, 0x2e,
	0xb6, 0xcd, 0x12, 0x8d, 0x8f, 0x21, 0x67, 0x11, 0xb2, 0x69, 0x63, 0xf6, 0x97, 0x63, 0x56, 0xdd,
	0x4e, 0x8b, 0x77, 0x61, 0xd3, 0x4e, 0x92, 0x47, 0x24, 0x6b, 0x87, 0x9c, 0x9f, 0x1e, 0xdb, 0x11,
	0xde, 0x81, 0x9c, 0x11, 0x7d, 0x96, 0x90, 0x82, 0x85, 0xdd, 0x01, 0x3f, 0x86, 0xfc, 0x8d, 0xee,
	0x0c, 0x55, 0x4c, 0x8a, 0x0e, 0xbe, 0xd1, 0x57, 0x2a, 0xc6, 0xa7, 0x50, 0x18, 0x30, 0x43, 0x23,
	0x6a, 0x28, 0xd9, 0xaa, 0x66, 0xeb, 0xa5, 0xe6, 0x73, 0x5f, 0xc9, 0xd0, 0xbf, 0xff, 0xae, 0xfe,
	0x87, 0x99, 0xae, 0x95, 0x18, 0x35, 0x0e, 0x52, 0x9b, 0xf7, 0x19, 0xca, 0x4b, 0x14, 0xae, 0x40,
	0xb6, 0xcf, 0xc6, 0xb3, 0xa7, 0x9f, 0xfe, 0xe2, 0x23, 0xc8, 0x8d, 0x68, 0x3c, 0x64, 0x24, 0x63,
	0x1b, 0xdc, 0xf1, 0xdd, 0xe6, 0xf9, 0xf3, 0xcd, 0xf3, 0x4f, 0x93, 0x71, 0xe0, 0x24, 0x6f, 0x33,
	0xaf, 0xd1, 0x59, 0x11, 0x36, 0x95, 0xbb, 0xb5, 0xb6, 0x07, 0xbb, 0x17, 0x5c, 0x9b, 0xd3, 0xd0,
	0xf0, 0xd1, 0xf2, 0x20, 0x6b, 0x6f, 0x80, 0xdc, 0xa7, 0xb4, 0x14, 0x89, 0x66, 0xf8, 0x10, 0x20,
	0x5d, 0x42, 0x4d, 0x50, 0x35, 0x5b, 0x2f, 0x06, 0xc5, 0xf9, 0x16, 0xea, 0xe6, 0x6f, 0x04, 0xdb,
	0xce, 0xd1, 0x4e, 0x0c, 0x53, 0x09, 0x8d, 0xf1, 0x7b, 0x28, 0x2d, 0x34, 0x8d, 0x77, 0xd7, 0x8c,
	0xc1, 0x7b, 0x94, 0xbe, 0xce, 0x3c, 0xe0, 0x5a, 0xd4, 0x60, 0x72, 0x8b, 0xf2, 0x15, 0x74, 0x82,
	0x5e, 0x21, 0xfc, 0x0d, 0x2a, 0xab, 0x65, 0xe1, 0x03, 0x9b, 0xb6, 0xa6, 0x11, 0xef, 0x70, 0x0d,
	0xeb, 0x7a, 0xa9, 0x15, 0x26, 0xb7, 0x68, 0xe3, 0x24, 0x53, 0x47, 0xcd, 0x5f, 0x08, 0xca, 0x8e,
	0x3c, 0xa7, 0x49, 0x14, 0x33, 0x85, 0xdb, 0xb0, 0x75, 0x25, 0x23, 0x6a, 0xd8, 0xa5, 0x51, 0x8c,
	0x0e, 0xf0, 0x41, 0x5a, 0xdd, 0x22, 0xfc, 0xcf, 0xda, 0xf3, 0x93, 0x5b, 0x94, 0xa9, 0x20, 0xdc,
	0x02, 0xb8, 0x34, 0x42, 0xce, 0x2a, 0xf6, 0x52, 0xe9, 0x1d, 0xf8, 0x3f, 0x31, 0x67, 0x2f, 0xbe,
	0x1f, 0xf5, 0xb8, 0xf9, 0x31, 0xec, 0xfa, 0xa1, 0x18, 0x34, 0x66, 0xc2, 0xf4, 0x2b, 0xfb, 0xbd,
	0x86, 0x66, 0x6a, 0xc4, 0x43, 0xd6, 0x50, 0x32, 0xec, 0xe6, 0xed, 0x0a, 0x1c, 0xff, 0x0d, 0x00,
	0x00, 0xff, 0xff, 0x18, 0x27, 0x54, 0x68, 0xa3, 0x04, 0x00, 0x00,
}