package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"codeberg.org/mjh/lrctl/grpc/shutdown"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"

	"codeberg.org/mjh/LibRate/cfg"
)

type GrpcServer struct {
	shutdown.UnimplementedShutdownServiceServer
	App    *fiber.App
	Log    *zerolog.Logger
	Config *cfg.GrpcConfig
}

// RunGrpcServer is the entry point for the GRPC server.
func RunGrpcServer(s *GrpcServer) {
	registerGRPC(s)
}

func registerGRPC(srv *GrpcServer) {
	listener := listenGRPC(srv)
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, os.Stdout, os.Stderr))

	s := grpc.NewServer()

	shutdown.RegisterShutdownServiceServer(s, srv)

	reflection.Register(s)

	err := s.Serve(listener)
	if err != nil {
		srv.Log.Fatal().Err(err).Msgf("failed to serve: %v", err)
	}
}

func listenGRPC(srv *GrpcServer) net.Listener {
	address := fmt.Sprintf("%s:%d", srv.Config.Host, srv.Config.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		srv.Log.Fatal().Err(err).Msgf("failed to listen: %v", err)
	}
	srv.Log.Info().Msgf("listening on %s", address)
	return listener
}

func (s *GrpcServer) SendShutdown(ctx context.Context, req *shutdown.ShutdownRequest) (*shutdown.ShutdownResponse, error) {
	if req.Message != "shutdown" {
		s.Log.Warn().Msg("shutdown requested with invalid message")
		return &shutdown.ShutdownResponse{Received: false}, nil
	}
	s.Log.Info().Msg("shutdown requested")

	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	if req.Timeout == nil && s.Config.ShutdownTimeout < 0 {
		if err := s.App.Shutdown(); err != nil {
			s.Log.Error().Err(err).Msg("failed to shutdown")
			return &shutdown.ShutdownResponse{Received: false}, err
		}
		return &shutdown.ShutdownResponse{Received: true}, nil
	}
	err := s.App.ShutdownWithContext(ctx)
	if err != nil {
		s.Log.Error().Err(err).Msg("failed to shutdown")
		return &shutdown.ShutdownResponse{Received: false}, err
	}
	s.Log.Info().Msg("shutdown complete")
	return &shutdown.ShutdownResponse{Received: true}, nil
}
