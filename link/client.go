package link

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/strotz/chainsaw/link/def"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

// TODO: Connected is not exactly correct API. Think about proper way to inform
// that the client is connected to the server. Maybe use a channel or a callback?

type Client struct {
	conn            *grpc.ClientConn
	chain           def.ChainClient
	Connected       atomic.Bool // true if connected to server
	queueIn         chan *def.Event
	AcceptedCounter atomic.Uint64
	SentCounter     atomic.Uint64
	ReceivedCounter atomic.Uint64
}

var ErrNotConnected = errors.New("not connected to server")

func NewClient() (*Client, error) {
	var opts []grpc.DialOption
	// TODO: implement TLS
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	c, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		return nil, err
	}
	chain := def.NewChainClient(c)
	return &Client{
		conn:    c,
		chain:   chain,
		queueIn: make(chan *def.Event, 100), // TODO: add a parameter for the queue length
	}, nil
}

func (c *Client) Run(ctx context.Context) error {
	if c.conn == nil {
		return ErrNotConnected
	}
	stream, err := c.chain.Do(ctx)
	if err != nil {
		return err
	}
	c.Connected.Store(true)
	slog.Debug("Client connected to server", "addr", *serverAddr)

	var wg sync.WaitGroup

	// Loop to receive messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			in, err := stream.Recv()
			// Check that higher context is not canceled
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err == io.EOF {
				log.Fatalln("Server closed the stream")
				return
			}
			if err != nil {
				log.Fatalln("Error receiving message:", err)
				return
			}
			// Process the received message
			slog.Debug("Client get message", "message", in)
			c.ReceivedCounter.Add(1)
		}
	}()

	// Loop to send messages.
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: should this error be handled?
		defer stream.CloseSend()
		for {
			select {
			// TODO: good?
			case in, good := <-c.queueIn:
				if !good {
					slog.Debug("Send closed by request")
					return
				}
				slog.Debug("Sending event", "in", in)
				err := stream.Send(in)
				if err == io.EOF {
					// TODO: retry? what is expected now: recreate clent or call Do again?
					log.Fatalln("Client disconnected from server")
					return
				}
				if err != nil {
					log.Fatalln("failed to send the message, error:", err)
				}
				c.SentCounter.Add(1)
			case <-ctx.Done():
				slog.Debug("Send cancelled by context")
				return
			}
		}
	}()

	// Wait for both send and receive to finish
	wg.Wait()
	return nil
}

func (c *Client) SendAndReceive(in *def.Event_StatusRequest, out *def.Event_StatusResponse) error {
	if c.chain == nil {
		return ErrNotConnected
	}
	msg := &def.Event{
		CallId:  &def.CallId{Id: "abc"},
		Payload: in,
	}
	c.queueIn <- msg
	c.AcceptedCounter.Add(1)
	return nil
}

func (c *Client) Close() error {
	c.Connected.Store(false)
	if c.conn != nil {
		err := c.conn.Close()
		c.chain = nil
		c.conn = nil
		return err
	}
	return nil
}
