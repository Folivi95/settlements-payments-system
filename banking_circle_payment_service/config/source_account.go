package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/adapters/env"
	"github.com/saltpay/settlements-payments-system/internal/projectpath"
)

func LoadSourceAccounts() ([]models.SourceAccount, error) {
	var sourceAccounts []models.SourceAccount

	envName := os.Getenv("ENV_NAME")
	if envName == "" {
		return sourceAccounts, errors.New("could not find ENV_NAME environment variable")
	}

	if envName == string(env.Tilt) {
		// use local source-accounts
		envName = string(env.Local)
	}

	// load env vars from `.envName.env`
	fileName := fmt.Sprintf("/banking_circle_payment_service/config/source-accounts.%s.json", envName)
	filePath := path.Join(projectpath.Root, fileName)

	byteValue, _ := ioutil.ReadFile(filePath)
	err := json.Unmarshal(byteValue, &sourceAccounts)
	if err != nil {
		return sourceAccounts, err
	}

	return sourceAccounts, nil
}
