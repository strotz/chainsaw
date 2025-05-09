package hello

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link"
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

	require.EqualError(t, c.Run(r.Ctx), "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp [::1]:50051: connect: connection refused\"")
}

// Validate that client and server can exchange messages
func TestRunHello(t *testing.T) {
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
		require.NoError(t, c.Run(r.Ctx))
	}()

	// Wait for the client to connect. It is necessary, to avoid error from c.Run()
	require.NoError(t, r.WaitFor(func() bool {
		return c.Connected.Load()
	}))
	slog.Debug("Connected")

	req := &def.Event_StatusRequest{}
	resp := &def.Event_StatusResponse{}
	require.NoError(t, c.SendAndReceive(req, resp))
	require.Equal(t, uint64(1), c.AcceptedCounter.Load())

	// Wait for the increment of sent counter.
	require.NoError(t, r.WaitFor(func() bool {
		return c.SentCounter.Load() > 0
	}))
	slog.Debug("Sent")

	require.NoError(t, r.WaitFor(func() bool {
		return c.ReceivedCounter.Load() > 0
	}))
}
