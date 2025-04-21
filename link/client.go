package link

import (
	"errors"

	"github.com/strotz/chainsaw/link/def"
)

type Client struct {
	transport def.ChainClient
}

var ErrNotConnected = errors.New("not connected to server")

func (c *Client) SendAndReceive(in *def.Event_StatusRequest, out *def.Event_StatusResponse) error {
	if c.transport == nil {
		return ErrNotConnected
	}
	return nil
}
