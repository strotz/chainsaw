package link

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"sync"

	"github.com/strotz/chainsaw/link/def"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

type Client struct {
	conn  *grpc.ClientConn
	chain def.ChainClient
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
		conn:  c,
		chain: chain,
	}, nil
}

func (c Client) Run(ctx context.Context) error {
	if c.conn == nil {
		return ErrNotConnected
	}
	stream, err := c.chain.Do(ctx)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup

	// Loop to receive messages
	wg.Add(1)
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// The server has closed the stream
				wg.Done()
				return
			}
			if err != nil {
				log.Fatalln("Error receiving message:", err)
			}
			// Process the received message
			log.Println("Received message:", in.String())
		}
	}()

	// Loop to send messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: should this error be handled?
		defer stream.CloseSend()
		select {
		case <-ctx.Done():
			return
			// TODO: send a message to the server
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
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.chain = nil
		c.conn = nil
		return err
	}
	return nil
}
