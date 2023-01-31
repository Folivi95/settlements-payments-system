package env

import "fmt"

// i put this in the adapters package as i feel our adapters are "aware" of
// environments and do different things accordingly

type EnvName string

const (
	Local                     EnvName = "local"
	Tilt                      EnvName = "tilt"
	IntegrationGlobalPlatform EnvName = "integration_gp"
	ProductionGlobalPlatform  EnvName = "production_gp"
	PipelineGlobalPlatform    EnvName = "infra_pipeline"
)

func (e *EnvName) Decode(value string) error {
	switch value {
	case "local":
		*e = Local
	case "tilt":
		*e = Tilt
	case "integration_gp":
		*e = IntegrationGlobalPlatform
	case "production_gp":
		*e = ProductionGlobalPlatform
	case "infra_pipeline":
		*e = PipelineGlobalPlatform
	default:
		return fmt.Errorf("unrecognised env name %q", value)
	}
	return nil
}
