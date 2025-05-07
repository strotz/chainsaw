package link

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/strotz/chainsaw/link/def"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
	queueIn         chan *def.Envelope
	AcceptedCounter atomic.Uint64
	SentCounter     atomic.Uint64
	ReceivedCounter atomic.Uint64
	RetryCounter    atomic.Uint64
	RetryDelay      time.Duration
	recipients      *table
}

var ErrNotInitialized = errors.New("connection not initialized")

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
		conn:       c,
		chain:      chain,
		queueIn:    make(chan *def.Envelope, 100), // TODO: add a parameter for the queue length
		RetryDelay: time.Second,
		recipients: newTable(),
	}, nil
}

// Start manages connection to the server including restore connection until context is canceled.
func (c *Client) Start(ctx context.Context) error {
	if c.chain == nil {
		return ErrNotInitialized
	}
	go c.recipients.run()
	for {
		c.RetryCounter.Add(1)
		err := c.processStream(ctx)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			slog.Debug("Stream error", "message", err)
			time.Sleep(c.RetryDelay)
		}
	}
}

func (c *Client) processStream(ctx context.Context) error {
	stream, err := c.chain.Do(ctx)
	if err != nil {
		return err
	}
	slog.Debug("Client called server", "addr", *serverAddr)
	c.Connected.Store(true)
	defer c.Connected.Store(false)

	readerError := make(chan error, 1)
	writerError := make(chan error, 1)
	stop := make(chan struct{})
	defer close(stop)

	// Loop to receive messages
	go func() {
		receiveLoop := func() error {
			for {
				in, err := stream.Recv()
				// Check that higher context is not canceled
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-stop:
					return nil
				default:
				}
				if err == io.EOF {
					slog.Debug("Server closed the stream")
					return err
				}
				// TODO: why it is not just EOF?
				if status.Code(err) == codes.Unavailable {
					slog.Debug("Server is unavailable", "err", err)
					return err
				}
				if err != nil {
					log.Fatalf("Error receiving message: %v", err)
				}
				// Process the received message
				slog.Debug("Client get message", "message", in)
				c.ReceivedCounter.Add(1)
				c.recipients.post(in.CallId.Id, in.GetEvent())
			}
		}
		err = receiveLoop()
		readerError <- err
	}()

	// Loop to send messages.
	go func() {
		writerLoop := func() error {
			defer slog.Debug("Exiting write loop")
			// TODO: should this error be handled?
			defer stream.CloseSend()
			for {
				select {
				case <-ctx.Done():
					slog.Debug("Send cancelled by context")
					return ctx.Err()
				case <-stop:
					return nil
				case in, good := <-c.queueIn:
					if !good {
						slog.Debug("Send closed by request")
						return errors.New("service closed")
					}
					slog.Debug("Sending event", "in", in)
					err := stream.Send(in)
					if err == io.EOF {
						slog.Debug("Client disconnected from server")
						return err
					}
					if err != nil {
						log.Fatalln("failed to send the message, error:", err)
					}
					c.SentCounter.Add(1)
				}
			}
		}
		err = writerLoop()
		writerError <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-readerError:
		return err
	case err := <-writerError:
		return err
	}
}

// SendAndReceive sends a message to the server and waits for a response.
func (c *Client) SendAndReceive(ctx context.Context, in *def.Event, out **def.Event) error {
	if c.chain == nil {
		return ErrNotInitialized
	}
	var err error
	*out, err = c.sendAndRecv(ctx, "random string", in)
	return err
}

func (c *Client) sendAndRecv(ctx context.Context, id string, in *def.Event) (*def.Event, error) {
	msg := &def.Envelope{
		CallId: &def.CallId{Id: id},
		Event:  in,
	}

	// TODO: how we are going to deal with multiple (idempotent messages). Pick first?
	sub := make(chan *def.Event, 1)
	c.recipients.addRecipient(id, sub)

	c.queueIn <- msg
	c.AcceptedCounter.Add(1)

	select {
	case <-ctx.Done():
		c.recipients.remove(id)
		return nil, ctx.Err()
	case out := <-sub:
		return out, nil
	}
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
