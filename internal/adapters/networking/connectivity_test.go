//go:build unit
// +build unit

package networking_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/adapters/networking"
)

func TestWaitForConnectivity(t *testing.T) {
	is := is.New(t)

	t.Run("errors when not connectable", func(t *testing.T) {
		ctx := context.Background()
		unreachableAddress := []string{"unreachable.address:11234"}

		err := networking.WaitForConnectivity(ctx, unreachableAddress, 2, 100*time.Millisecond)

		is.True(err != nil)
	})

	t.Run("should error if at least one address fails", func(t *testing.T) {
		ctx := context.Background()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer server.Close()

		addresses := []string{server.Listener.Addr().String(), "unreachable.address:11234"}
		err := networking.WaitForConnectivity(ctx, addresses, 2, 100*time.Millisecond)
		is.True(err != nil)
	})

	t.Run("does not error when connectable", func(t *testing.T) {
		ctx := context.Background()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer server.Close()

		addresses := []string{server.Listener.Addr().String(), server.Listener.Addr().String()}
		err := networking.WaitForConnectivity(ctx, addresses, 2, 100*time.Millisecond)

		is.NoErr(err)
	})
}
