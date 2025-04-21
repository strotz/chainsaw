package link

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link/def"
	"go.uber.org/mock/gomock"
)

var _ def.ChainClient = (*MockChainClient)(nil)

func TestCompileClient(t *testing.T) {
	cli := &Client{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := NewMockChainClient(ctrl)
	cli.transport = c

	in := &def.Event_StatusRequest{}
	out := &def.Event_StatusResponse{}
	require.NoError(t, cli.SendAndReceive(in, out))
}
