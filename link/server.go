package link

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"sync/atomic"

	"github.com/strotz/chainsaw/link/def"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type Server struct {
	x def.ChainServer
}

type imp struct {
	received atomic.Int64
	sent     atomic.Int64
}

func (i *imp) Do(server def.Chain_DoServer) error {
	for {
		in, err := server.Recv()
		if err == io.EOF {
			slog.Debug("Client closed connection with server")
			return nil
		}
		if err != nil {
			slog.Warn("Error with server connection", "error", err)
			return err
		}
		slog.Debug("Received by server", "in", in)
		i.received.Add(1)
		// TODO: this is dirty hack to make hello test meaningful
		if x := in.GetEvent().GetStatusRequest(); x != nil {
			y := def.MakeEnvelope(in.CallId.Id,
				&def.Event_StatusResponse{
					StatusResponse: &def.StatusResponse{
						ReceivedMessagesCounter: i.received.Load(),
						SentMessagesCounter:     i.sent.Load(),
					},
				},
			)
			if err := server.Send(y); err != nil {
				return err
			}
			i.sent.Add(1)
		}
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	// Borrowed from grpc-go examples
	//if *tls {
	//	if *certFile == "" {
	//		*certFile = data.Path("x509/server_cert.pem")
	//	}
	//	if *keyFile == "" {
	//		*keyFile = data.Path("x509/server_key.pem")
	//	}
	//	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	//	if err != nil {
	//		log.Fatalf("Failed to generate credentials: %v", err)
	//	}
	//	opts = []grpc.ServerOption{grpc.Creds(creds)}
	//}
	grpcServer := grpc.NewServer(opts...)
	s.x = &imp{}
	def.RegisterChainServer(grpcServer, s.x)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return nil
}
