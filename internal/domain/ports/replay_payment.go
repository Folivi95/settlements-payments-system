//go:generate moq -out mocks/replay_payment_moq.go -pkg=mocks . ReplayPayment

package ports

import "context"

type ReplayPayment interface {
	Execute(ctx context.Context, currency string, file string) error
}
