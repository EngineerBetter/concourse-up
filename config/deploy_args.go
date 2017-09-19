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
	DBSize      string
	// DBSizeIsSet is true if the user has manually specified the db-size (ie, it's not the default)
	DBSizeIsSet bool
}

// WorkerSizes are the permitted concourse worker sizes
var WorkerSizes = []string{"medium", "large", "xlarge"}

// DBSizes maps SML sizes to RDS instance classes
var DBSizes = map[string]string{
	"small":  "db.t2.small",
	"medium": "db.t2.medium",
	"large":  "db.m4.large",
}

// Validate validates that flag interdependencies
func (args DeployArgs) Validate() error {
	if err := args.validateCertFields(); err != nil {
		return err
	}

	if err := args.validateWorkerFields(); err != nil {
		return err
	}

	if err := args.validateDBFields(); err != nil {
		return err
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

	return nil
}

func (args DeployArgs) validateWorkerFields() error {
	if args.WorkerCount < 1 {
		return errors.New("minimum of workers is 1")
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

func (args DeployArgs) validateDBFields() error {
	if _, ok := DBSizes[args.DBSize]; !ok {
		return fmt.Errorf("unknown DB size: `%s`. Valid sizes are: %v", args.DBSize, DBSizes)
	}

	return nil
}
