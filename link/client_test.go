package link

import (
	"context"
	"io"
	"log"
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
	go cli.recipients.run()

	s := sim.NewSequenceSim()
	s.Add(sim.ClientSend, def.MakeEnvelope("test", &def.Event_StatusRequest{}))
	s.Add(sim.ClientRecv, def.MakeEnvelope("test", &def.Event_StatusResponse{}))
	cli.chain = s

	// TODO: it should be another set of test to cover Start function

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := cli.processStream(context.TODO()); err != io.EOF {
			log.Fatalf("expected EOF, got %v", err)
		}
	}()

	out, err := cli.sendAndRecv(context.TODO(), "test", def.MakeEvent(&def.Event_StatusRequest{}))
	require.NoError(t, err)
	require.True(t, proto.Equal(out, def.MakeEvent(&def.Event_StatusResponse{})), "out: %v", out)

	wg.Wait()

	require.True(t, s.IsDone())
	cli.recipients.stop()
}
