package sim

import (
	"context"
	"io"
	"log"
	"sync"

	"github.com/strotz/chainsaw/link/def"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// sequenceSim is an apparatus to simulate the sequence of events

type Direction int

const (
	// ClientSend - client expected to send item
	ClientSend Direction = iota
	// ClientRecv - client expected to receive item
	ClientRecv
)

type item struct {
	d Direction
	v *def.Envelope
}

type zeroClientStream struct{}

func (s *zeroClientStream) Header() (metadata.MD, error) {
	//TODO implement me
	panic("implement me")
}

func (s *zeroClientStream) Trailer() metadata.MD {
	//TODO implement me
	panic("implement me")
}

func (s *zeroClientStream) CloseSend() error {
	return nil
}

func (s *zeroClientStream) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (s *zeroClientStream) SendMsg(m any) error {
	//TODO implement me
	panic("implement me")
}

func (s *zeroClientStream) RecvMsg(m any) error {
	//TODO implement me
	panic("implement me")
}

type sequenceSim struct {
	cond  sync.Cond
	items []*item
	zeroClientStream
}

type SequenceSim interface {
	def.ChainClient
	Add(direction Direction, v *def.Envelope)
	IsDone() bool
}

func NewSequenceSim() SequenceSim {
	return &sequenceSim{
		cond: sync.Cond{L: &sync.Mutex{}},
	}
}

func (s *sequenceSim) Do(ctx context.Context, opts ...grpc.CallOption) (def.Chain_DoClient, error) {
	return s, nil
}

func (s *sequenceSim) Send(e *def.Envelope) error {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	if len(s.items) == 0 {
		log.Fatalf("No items to send")
	}
	x := s.items[0]
	if x.d != ClientSend {
		log.Fatal("Expected ClientSend, got", x.d)
	}
	if !proto.Equal(e, x.v) {
		log.Fatalf("Sent: %v, Expected: %v", e, x.v)
	}
	s.items = s.items[1:]
	s.cond.Broadcast()
	return nil
}

func (s *sequenceSim) Recv() (*def.Envelope, error) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	for {
		if len(s.items) == 0 {
			return nil, io.EOF
		}
		x := s.items[0]
		// Recv is a blocking call and can pass ClientSend items
		if x.d != ClientRecv {
			s.cond.Wait()
			continue
		}
		s.items = s.items[1:]
		s.cond.Broadcast()
		return x.v, nil
	}
}

func (s *sequenceSim) Add(direction Direction, e *def.Envelope) {
	s.items = append(s.items, &item{d: direction, v: e})
}

func (s *sequenceSim) IsDone() bool {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	return len(s.items) == 0
}

func (s *sequenceSim) top() *item {
	if len(s.items) == 0 {
		return nil
	} else {
		return s.items[0]
	}
}
