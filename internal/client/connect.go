package client

import (
	"context"
	"fmt"
	"time"

	retry "github.com/avast/retry-go/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
)

func ConnectToService(ctx context.Context, name, port string) (*grpc.ClientConn, error) {
	// set timeout for connection to 15 seconds
	var (
		conn *grpc.ClientConn
		err  error
	)
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("failed to connect to %s: %v", name, ctx.Err())
	default:
		err = retry.Do(
			func() error {
				dialOpts := append([]grpc.DialOption{},
					grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
				)
				conn, err = grpc.DialContext(ctx, fmt.Sprintf("%s:%s", name, port), dialOpts...)
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
}
