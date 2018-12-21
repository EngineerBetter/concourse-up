package fly

import "github.com/EngineerBetter/concourse-up/config"

// Pipeline is interface for self update pipeline
type Pipeline interface {
	BuildPipelineParams(config config.Config) (Pipeline, error)
	GetConfigTemplate() string
}
