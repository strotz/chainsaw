package hello

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link"
	"github.com/strotz/chainsaw/link/clientfixture"
	"github.com/strotz/chainsaw/link/def"
	"github.com/strotz/chainsaw/link/serverfixture"
	"github.com/strotz/chainsaw/link/tests"
)

// Validate that client is not happy when the server is not up
func TestOnlyClient(t *testing.T) {
	r := tests.Setup(t).WithTimeout(5 * time.Second)
	defer r.Close()

	c, err := link.NewClient()
	require.NoError(t, err)
	defer c.Close()

	c.RetryDelay = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(r.Ctx, 3*time.Second)
	defer cancel()
	require.Errorf(t, c.Start(ctx), "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp [::1]:50051: connect: connection refused\"")
	require.EqualValues(t, 31, int(c.RetryCounter.Load()))
}

// Validate that client and server can exchange messages
func TestRunHello(t *testing.T) {
	r := tests.Setup(t).WithTimeout(5 * time.Second)
	defer r.Close()

	s := serverfixture.Fixture{}
	require.NoError(t, s.StartServer(r.Ctx, &r.WaitDone))

	cf := clientfixture.Fixture{}
	err := cf.RunConnected(r, t)
	require.NoError(t, err)
	defer cf.Close()

	req := def.MakeEvent(&def.Event_StatusRequest{})
	resp := def.MakeEvent(&def.Event_StatusResponse{})
	require.NoError(t, cf.Client.SendAndReceive(r.Ctx, req, &resp))
	require.Equal(t, uint64(1), cf.Client.AcceptedCounter.Load())

	// Wait for the increment of sent counter.
	require.NoError(t, r.WaitFor(func() bool {
		return cf.Client.SentCounter.Load() > 0
	}))
	slog.Debug("Sent")

	require.NoError(t, r.WaitFor(func() bool {
		return cf.Client.ReceivedCounter.Load() > 0
	}))

	slog.Debug("Received", "resp", resp)
	require.NotNil(t, resp.GetStatusResponse())
	require.Equal(t, int64(1), resp.GetStatusResponse().ReceivedMessagesCounter)
	require.Equal(t, int64(0), resp.GetStatusResponse().SentMessagesCounter)
}

func TestRetry(t *testing.T) {
	r := tests.Setup(t).WithTimeout(5 * time.Second)
	defer r.Close()

	s := serverfixture.Fixture{}
	require.NoError(t, s.StartServer(r.Ctx, &r.WaitDone))

	c, err := link.NewClient()
	require.NoError(t, err)
	defer c.Close()

	r.WaitDone.Add(1)
	go func() {
		defer r.WaitDone.Done()
		require.ErrorIs(t, context.Canceled, c.Start(r.Ctx))
	}()

	//Wait for the client to connect. It is necessary, to avoid error from c.Start()
	require.NoError(t, r.WaitFor(func() bool {
		return c.Connected.Load()
	}))
	slog.Debug("Connected")

	s.Server.Kill()

	//Client should indicate that it is disconnected.
	require.NoError(t, r.WaitFor(func() bool {
		return !c.Connected.Load()
	}))
	slog.Debug("Disconnected")

	require.NoError(t, s.StartServer(r.Ctx, &r.WaitDone))
	require.NoError(t, r.WaitFor(func() bool {
		return c.Connected.Load()
	}))
	slog.Debug("Back online")

	req := def.MakeEvent(&def.Event_StatusRequest{})
	resp := def.MakeEvent(&def.Event_StatusResponse{})
	require.NoError(t, c.SendAndReceive(r.Ctx, req, &resp))
	require.Equal(t, uint64(1), c.AcceptedCounter.Load())

	// Wait for the increment of sent counter.
	require.NoError(t, r.WaitFor(func() bool {
		return c.SentCounter.Load() == 1
	}))
	slog.Debug("Sent")

	require.NoError(t, r.WaitFor(func() bool {
		return c.ReceivedCounter.Load() == 1
	}))
	slog.Debug("Received")
}

// Cancel request while it is waiting for the response via context.
func TestRequestTimeout(t *testing.T) {
	r := tests.Setup(t).WithTimeout(5 * time.Second)
	defer r.Close()

	s := serverfixture.Fixture{}
	require.NoError(t, s.StartServer(r.Ctx, &r.WaitDone))

	cf := clientfixture.Fixture{}
	err := cf.RunConnected(r, t)
	require.NoError(t, err)
	defer cf.Close()

	req := def.MakeEvent(&def.Event_StatusRequest{})
	resp := def.MakeEvent(&def.Event_NoOp{})
	ctx, cancel := context.WithTimeout(r.Ctx, time.Second)
	defer cancel()
	require.NoError(t, cf.Client.SendAndReceive(ctx, req, &resp))
}
