package client

import (
	"context"
	"fmt"
	"time"

	retry "github.com/avast/retry-go/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectToService(ctx context.Context, name, port string) (*grpc.ClientConn, error) {
	// set timeout for connection to 15 seconds
	var conn *grpc.ClientConn
	var err error

	err = retry.Do(
		func() error {
			conn, err = grpc.Dial(name+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return fmt.Errorf("failed to connect to %s: %v", name, err)
			}
			return nil
		},
		retry.Attempts(6),                  // 3 retries per second for 15 seconds
		retry.Delay(2500*time.Millisecond), // Delay between retries
		retry.Context(ctx),                 // Pass the context to be used for cancellation
	)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s after retries: %v", name, err)
	}

	return conn, nil
}
