package networking

import (
	"context"
	"fmt"
	"net"
	"time"

	zapctx "github.com/saltpay/go-zap-ctx"
)

// WaitForConnectivity will iterate the addresses supplied and retry / wait until network is reachable
// before returning ok.
func WaitForConnectivity(
	ctx context.Context,
	remoteAddrToWaitFor []string,
	retryCount int,
	retryDelay time.Duration,
) error {
	// Check each address one by one
	for _, address := range remoteAddrToWaitFor {
		start := time.Now()
		for i := 0; i <= retryCount; i++ {
			// if we tested the number of times we expected it, we give up
			if i == retryCount {
				return fmt.Errorf("could not connect to %s; tried %d times in %v", address, retryCount, time.Since(start))
			}

			zapctx.Info(ctx, fmt.Sprintf("attempt %d in connecting to %s", i, address))
			conn, err := net.Dial("tcp", address)
			if err == nil {
				// No error, managed to connect. We can skip this address
				_ = conn.Close()
				break
			}

			// delay retry and try again
			time.Sleep(retryDelay)
		}
	}

	return nil
}
