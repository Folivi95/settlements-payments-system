package ports

import "context"

type SecretsRepo interface {
	GetSecret(ctx context.Context, name string) (string, error)
	LookupSecret(name string) (string, bool)
}
