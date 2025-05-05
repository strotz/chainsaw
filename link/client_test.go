package link

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link/def"
	"github.com/strotz/chainsaw/link/sim"
	"google.golang.org/protobuf/proto"
)

func TestCompileClient(t *testing.T) {
	initTest()

	cli := &Client{
		queueIn:    make(chan *def.Envelope, 1), // To unblock SendAndReceive
		recipients: newTable(),
	}
	// TODO: start in real code
	go cli.recipients.run()

	s := sim.NewSequenceSim(t)
	s.Add(sim.ClientSend, def.MakeEnvelope("test", &def.Event_StatusRequest{}))
	s.Add(sim.ClientRecv, def.MakeEnvelope("test", &def.Event_StatusResponse{}))
	cli.chain = s

	// TODO: it should be another set of test to cover Start function

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.ErrorIs(t, cli.processStream(context.TODO()), io.EOF)
	}()

	out, err := cli.sendAndRecv(context.TODO(), "test", def.MakeEvent(&def.Event_StatusRequest{}))
	require.NoError(t, err)
	require.True(t, proto.Equal(out, def.MakeEvent(&def.Event_StatusResponse{})), "out: %v", out)

	wg.Wait()

	require.True(t, s.IsDone())
	cli.recipients.stop()
}
