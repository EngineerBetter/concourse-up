package config

import (
	"errors"
	"fmt"
)

// DeployArgs are arguments passed to the deploy command
type DeployArgs struct {
	IAAS        string
	AWSRegion   string
	Domain      string
	TLSCert     string
	TLSKey      string
	WorkerCount int
	WorkerSize  string
	SelfUpdate  bool
}

// WorkerSizes are the permitted concourse worker sizes
var WorkerSizes = []string{"medium", "large", "xlarge"}

// Validate validates that flag interdependencies
func (args DeployArgs) Validate() error {
	err := args.validateCertFields()
	if err != nil {
		return err
	}

	knownSize := false
	for _, size := range WorkerSizes {
		if args.WorkerSize == size {
			knownSize = true
		}
	}

	if !knownSize {
		return fmt.Errorf("unknown worker size: `%s`. Valid sizes are: %v", args.WorkerSize, WorkerSizes)
	}

	return nil
}

func (args DeployArgs) validateCertFields() error {
	if args.TLSKey != "" && args.TLSCert == "" {
		return errors.New("--tls-key requires --tls-cert to also be provided")
	}
	if args.TLSCert != "" && args.TLSKey == "" {
		return errors.New("--tls-cert requires --tls-key to also be provided")
	}
	if (args.TLSKey != "" || args.TLSCert != "") && args.Domain == "" {
		return errors.New("custom certificates require --domain to be provided")
	}
	if args.WorkerCount < 1 {
		return errors.New("minimum of workers is 1")
	}

	return nil
}
