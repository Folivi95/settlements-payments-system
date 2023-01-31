package secretsmanager

import (
	"context"
	"fmt"
	"os"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"
)

type SecretsRepo struct{}

func NewSecretsRepo() *SecretsRepo {
	return &SecretsRepo{}
}

func (s SecretsRepo) GetSecret(ctx context.Context, name string) (string, error) {
	secret := os.Getenv(name)
	if secret == "" {
		return "", fmt.Errorf("missing secret from environment variable: %s", name)
	}

	zapctx.Info(ctx, "secret successfully read", zap.String("secret_name", name))
	return secret, nil
}

func (s SecretsRepo) LookupSecret(name string) (string, bool) {
	return os.LookupEnv(name)
}
