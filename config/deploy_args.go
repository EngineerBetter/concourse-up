package config

import (
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/urfave/cli.v1"
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
	WebSize     string
	SelfUpdate  bool
	DBSize      string
	// DBSizeIsSet is true if the user has manually specified the db-size (ie, it's not the default)
	DBSizeIsSet            bool
	AllowIPs               string
	GithubAuthClientID     string
	GithubAuthClientSecret string
	// GithubAuthIsSet is true if the user has specified both the --github-auth-client-secret and --github-auth-client-id flags
	GithubAuthIsSet bool
	Tags            cli.StringSlice
}

// WorkerSizes are the permitted concourse worker sizes
var WorkerSizes = []string{"medium", "large", "xlarge", "2xlarge", "4xlarge", "10xlarge", "16xlarge"}

// WebSizes are the permitted concourse web sizes
var WebSizes = []string{"small", "medium", "large", "xlarge", "2xlarge"}

// DBSizes maps SML sizes to RDS instance classes
var DBSizes = map[string]string{
	"small":   "db.t2.small",
	"medium":  "db.t2.medium",
	"large":   "db.m4.large",
	"xlarge":  "db.m4.xlarge",
	"2xlarge": "db.m4.2xlarge",
	"4xlarge": "db.m4.4xlarge",
}

// ModifyGithub allows mutation of github related fields
func (args *DeployArgs) ModifyGithub(GithubAuthClientID, GithubAuthClientSecret string, GithubAuthIsSet bool) {
	args.GithubAuthClientID = GithubAuthClientID
	args.GithubAuthClientSecret = GithubAuthClientSecret
	args.GithubAuthIsSet = GithubAuthIsSet
}

// Validate validates that flag interdependencies
func (args DeployArgs) Validate() error {
	if err := args.validateCertFields(); err != nil {
		return err
	}

	if err := args.validateWorkerFields(); err != nil {
		return err
	}

	if err := args.validateWebFields(); err != nil {
		return err
	}

	if err := args.validateDBFields(); err != nil {
		return err
	}

	if err := args.validateGithubFields(); err != nil {
		return err
	}

	if err := args.validateTags(); err != nil {
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
		return errors.New("minimum number of workers is 1")
	}

	for _, size := range WorkerSizes {
		if size == args.WorkerSize {
			return nil
		}
	}
	return fmt.Errorf("unknown worker size: `%s`. Valid sizes are: %v", args.WorkerSize, WorkerSizes)
}

func (args DeployArgs) validateWebFields() error {
	for _, size := range WebSizes {
		if size == args.WebSize {
			return nil
		}
	}
	return fmt.Errorf("unknown web node size: `%s`. Valid sizes are: %v", args.WebSize, WebSizes)
}

func (args DeployArgs) validateDBFields() error {
	if _, ok := DBSizes[args.DBSize]; !ok {
		return fmt.Errorf("unknown DB size: `%s`. Valid sizes are: %v", args.DBSize, DBSizes)
	}

	return nil
}

func (args DeployArgs) validateGithubFields() error {
	if args.GithubAuthClientID != "" && args.GithubAuthClientSecret == "" {
		return errors.New("--github-auth-client-id requires --github-auth-client-secret to also be provided")
	}
	if args.GithubAuthClientID == "" && args.GithubAuthClientSecret != "" {
		return errors.New("--github-auth-client-secret requires --github-auth-client-id to also be provided")
	}

	return nil
}

func (args DeployArgs) validateTags() error {
	for _, tag := range args.Tags {
		m, err := regexp.MatchString(`\w+=\w+`, tag)
		if err != nil {
			return err
		}
		if !m {
			return fmt.Errorf("`%v` is not in the format `key=value`", tag)
		}
	}
	return nil
}
