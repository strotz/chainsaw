package sim

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link/def"
)

func initTest() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
}

var _ def.ChainClient = &sequenceSim{}

func TestCreate(t *testing.T) {
	initTest()
	s := NewSequenceSim(t)
	require.True(t, s.IsDone())
}

func TestAddClientSend(t *testing.T) {
	initTest()
	s := NewSequenceSim(t)
	s.Add(ClientSend, &def.Envelope{})
	require.False(t, s.IsDone())
}

func TestPlayClientSend(t *testing.T) {
	initTest()
	s := NewSequenceSim(t)
	s.Add(ClientSend, &def.Envelope{})

	stream, err := s.Do(context.TODO())
	require.NoError(t, err)

	require.NoError(t, stream.Send(&def.Envelope{}))
	require.True(t, s.IsDone())
}

func TestPlayClientSendFailed(t *testing.T) {
	t.Skip("intentionally failed test to demonstrate not expected Send")

	initTest()
	s := NewSequenceSim(t)
	s.Add(ClientSend, &def.Envelope{})

	stream, err := s.Do(context.TODO())
	require.NoError(t, err)

	require.NoError(t, stream.Send(&def.Envelope{
		Event: &def.Event{},
	}))
	require.True(t, s.IsDone())
}

func TestPlayClientRecv(t *testing.T) {
	initTest()
	s := NewSequenceSim(t)
	s.Add(ClientRecv, &def.Envelope{
		CallId: &def.CallId{
			Id: "test",
		},
	})
	require.False(t, s.IsDone())

	stream, err := s.Do(context.TODO())
	require.NoError(t, err)

	x, err := stream.Recv()
	require.NoError(t, err)
	require.EqualValues(t, &def.Envelope{
		CallId: &def.CallId{
			Id: "test",
		},
	}, x)
	require.True(t, s.IsDone())
}

func TestPlayClientRecvAfterSend(t *testing.T) {
	initTest()
	s := NewSequenceSim(t)
	s.Add(ClientSend, &def.Envelope{
		CallId: &def.CallId{
			Id: "test_send",
		},
	})
	s.Add(ClientRecv, &def.Envelope{
		CallId: &def.CallId{
			Id: "test_recv",
		},
	})

	stream, err := s.Do(context.TODO())
	require.NoError(t, err)

	require.NoError(t, stream.Send(&def.Envelope{
		CallId: &def.CallId{
			Id: "test_send",
		},
	}))
	x, err := stream.Recv()
	require.NoError(t, err)
	require.EqualValues(t, &def.Envelope{
		CallId: &def.CallId{
			Id: "test_recv",
		},
	}, x)
	require.True(t, s.IsDone())
}
