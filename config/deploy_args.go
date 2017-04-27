package config

import "errors"

type DeployArgs struct {
	AWSRegion   string
	Domain      string
	TLSCert     string
	TLSKey      string
	WorkerCount int
}

// Validate validates that flag interdependencies
func (args DeployArgs) Validate() error {
	if args.TLSKey != "" && args.TLSCert == "" {
		return errors.New("--tls-key requires --tls-cert to also be provided")
	}
	if args.TLSCert != "" && args.TLSKey == "" {
		return errors.New("--tls-cert requires --tls-key to also be provided")
	}
	if (args.TLSKey != "" || args.TLSCert != "") && args.Domain == "" {
		return errors.New("custom certificates require --domain to be provided")
	}
	if args.WorkerCount <= 1 {
		return errors.New("minimum of workers is 1")
	}
	return nil
}
