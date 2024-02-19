package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	protodb "codeberg.org/mjh/lrctl/grpc/db"
	protosearch "codeberg.org/mjh/lrctl/grpc/search"
	"codeberg.org/mjh/lrctl/grpc/shutdown"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers/search"
	"codeberg.org/mjh/LibRate/controllers/search/meili"
	"codeberg.org/mjh/LibRate/db"
	searchdb "codeberg.org/mjh/LibRate/models/search"
)

type GrpcServer struct {
	shutdown.UnimplementedShutdownServiceServer
	protodb.UnimplementedDBServer
	protosearch.UnimplementedSearchServer
	Cache  *redis.Storage
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
	protodb.RegisterDBServer(s, srv)
	protosearch.RegisterSearchServer(s, srv)

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

func (s *GrpcServer) Init(ctx context.Context, req *protodb.InitRequest) (*protodb.InitResponse, error) {
	s.Log.Info().Msg("database initialization request received")

	ssl := *req.Ssl
	if req.Ssl == nil {
		ssl = "disable"
	}

	dsn := cfg.DBConfig{
		Engine:   req.Engine,
		Host:     req.Host,
		Port:     uint16(req.Port),
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		SSL:      ssl,
	}

	s.Log.Debug().Msgf("Initialization request parameters: %+v", req)

	if err := db.InitDB(&dsn, s.Log); err != nil {
		s.Log.Error().Err(err).Msg("failed to initialize database")
		return &protodb.InitResponse{Success: false}, err
	}
	s.Log.Info().Msg("database initialized")
	return &protodb.InitResponse{Success: true}, nil
}

func (s *GrpcServer) Migrate(ctx context.Context, req *protodb.MigrateRequest) (*protodb.MigrateResponse, error) {
	s.Log.Info().Msg("database migration request received")

	ssl := *req.Dsn.Ssl
	if req.Dsn.Ssl == nil {
		ssl = "disable"
	}

	dsn := cfg.DBConfig{
		Engine:         req.Dsn.Engine,
		Host:           req.Dsn.Host,
		Port:           uint16(req.Dsn.Port),
		User:           req.Dsn.User,
		Password:       req.Dsn.Password,
		Database:       req.Dsn.Database,
		SSL:            ssl,
		MigrationsPath: "/app/data/migrations",
	}
	conf := cfg.Config{
		DBConfig: dsn,
	}

	s.Log.Debug().Msgf("Migration request parameters: %+v", req)

	switch {
	case len(req.Migrations) == 0 || *req.All:
		if err := db.Migrate(s.Log, &conf); err != nil {
			return &protodb.MigrateResponse{
				Success: false,
				Errors: []*protodb.MigrationError{
					{
						MigrationPath: "migrations", // TODO: add more precise extraction of exception path
						Message:       err.Error(),
					},
				},
			}, err
		}
		return &protodb.MigrateResponse{
			Success: true,
			Errors:  nil,
		}, nil
	default:
		count := len(req.Migrations)
		for i, migration := range req.Migrations {
			if err := db.Migrate(s.Log, &conf, migration); err != nil {
				if req.Hardfail {
					if i < count {
						s.Log.Warn().Msgf("error while running migration at %s: %v", migration, err)
					} else {
						goto fail
					}
				} else {
					goto fail
				}
			fail:
				return &protodb.MigrateResponse{
					Success: false,
					Errors: []*protodb.MigrationError{
						{
							MigrationPath: migration,
							Message:       err.Error(),
						},
					},
				}, err
			}
		}
	}

	s.Log.Info().Msg("database migrated")
	return &protodb.MigrateResponse{Success: true}, nil
}

func (s *GrpcServer) BuildIndex(
	ctx context.Context,
	req *protosearch.BuildRequest,
) (res *protosearch.BuildResponse, err error) {
	conf := cfg.SearchConfig{
		Provider:      req.Config.Provider,
		MeiliHost:     *req.Config.MeiliHost,
		MeiliPort:     int(*req.Config.MeiliPort),
		CouchDBHost:   req.Config.Host,
		Port:          int(req.Config.Port),
		User:          req.Config.User,
		Password:      req.Config.Password,
		MainIndexPath: req.Config.IndexPath,
	}
	storage, err := searchdb.Connect(ctx, &conf, s.Log)
	if err != nil {
		return nil, err
	}

	switch req.Config.Provider {
	case "bleve":
		svc := search.NewService(ctx, nil, storage, req.Config.IndexPath, s.Cache, s.Log).OrElse(
			search.ServiceNoIndex(nil, storage, s.Cache, s.Log),
		)

		err = svc.CreateIndex(ctx, req.RuntimeStats, req.Config.IndexPath)
		if err != nil {
			return nil, err
		}
		return &protosearch.BuildResponse{
			DocumentCount:   1,
			TimePerDocument: 0.1,
		}, nil
	case "meilisearch", "meili":
		svc, err := meili.Connect(&conf, s.Log, nil, storage)
		if err != nil {
			return nil, err
		}

		if err = svc.CreateAllIndexes(ctx); err != nil {
			return nil, err
		}

		return &protosearch.BuildResponse{
			DocumentCount:   1,
			TimePerDocument: 0.1,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported search provider: %s", req.Config.Provider)
	}
}
