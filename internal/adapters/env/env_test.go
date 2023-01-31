//go:build unit
// +build unit

package env

import (
	"fmt"
	"testing"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

func TestEnvName_Decode(t *testing.T) {
	cases := []struct {
		Input       string
		ExpectedEnv EnvName
	}{
		{"local", Local},
		{"integration_gp", IntegrationGlobalPlatform},
		{"infra_pipeline", PipelineGlobalPlatform},
	}

	for _, testcase := range cases {
		t.Run(fmt.Sprintf("%q->%v", testcase.Input, testcase.ExpectedEnv), func(t *testing.T) {
			is := is.New(t)
			var value EnvName
			is.NoErr(value.Decode(testcase.Input))
			is.Equal(value, testcase.ExpectedEnv)
		})
	}

	t.Run("unrecongised value returns an error", func(t *testing.T) {
		is := is.New(t)
		var value EnvName
		err := value.Decode(testhelpers.RandomString())
		is.True(err != nil)
	})
}
