package cmd

import (
	"context"
	"testing"

	"codeberg.org/mjh/LibRate/cfg"
	protodb "codeberg.org/mjh/lrctl/grpc/db"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	l := zerolog.Nop()
	mockGrpcServer := GrpcServer{
		App: fiber.New(fiber.Config{}),
		Log: &l,
		Config: &cfg.GrpcConfig{
			Host:            "localhost",
			Port:            3030,
			ShutdownTimeout: 10,
		},
	}

	RunGrpcServer(&mockGrpcServer)

	req := protodb.InitRequest{
		Engine:   "postgres",
		Host:     "librate-db",
		Port:     uint32(5432),
		User:     "postgres",
		Password: "postgres",
		Database: "librate_test",
	}

	res, err := mockGrpcServer.Init(context.Background(), &req)
	assert.Equal(t, res.Success, true)
	assert.Nil(t, err)
}
