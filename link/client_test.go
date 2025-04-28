package link

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link/def"
	"go.uber.org/mock/gomock"
)

var _ def.ChainClient = (*MockChainClient)(nil)

func TestCompileClient(t *testing.T) {
	cli := &Client{
		queueIn: make(chan *def.Event, 1), // To unblock SendAndReceive
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// TODO: it should be another set of test to cover Run function
	c := NewMockChainClient(ctrl)
	cli.chain = c

	in := &def.Event_StatusRequest{}
	out := &def.Event_StatusResponse{}
	require.NoError(t, cli.SendAndReceive(in, out))
}
