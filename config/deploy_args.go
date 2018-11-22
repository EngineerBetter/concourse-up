package config

import (
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/urfave/cli.v1"
)

// DeployArgs are arguments passed to the deploy command
type DeployArgs struct {
	IAAS             string
	IAASIsSet        bool
	AWSRegion        string
	AWSRegionIsSet   bool
	Domain           string
	DomainIsSet      bool
	TLSCert          string
	TLSCertIsSet     bool
	TLSKey           string
	TLSKeyIsSet      bool
	WorkerCount      int
	WorkerCountIsSet bool
	WorkerSize       string
	WorkerSizeIsSet  bool
	WebSize          string
	WebSizeIsSet     bool
	SelfUpdate       bool
	SelfUpdateIsSet  bool
	DBSize           string
	// DBSizeIsSet is true if the user has manually specified the db-size (ie, it's not the default)
	DBSizeIsSet                 bool
	Namespace                   string
	NamespaceIsSet              bool
	AllowIPs                    string
	AllowIPsIsSet               bool
	GithubAuthClientID          string
	GithubAuthClientIDIsSet     bool
	GithubAuthClientSecret      string
	GithubAuthClientSecretIsSet bool
	// GithubAuthIsSet is true if the user has specified both the --github-auth-client-secret and --github-auth-client-id flags
	GithubAuthIsSet bool
	Tags            cli.StringSlice
	// TagsIsSet is true if the user has specified tags using --tags
	TagsIsSet       bool
	Spot            bool
	SpotIsSet       bool
	Zone            string
	ZoneIsSet       bool
	WorkerType      string
	WorkerTypeIsSet bool
}

// MarkSetFlags is marking the IsSet DeployArgs
func (a *DeployArgs) MarkSetFlags(c *cli.Context) error {
	for _, f := range c.FlagNames() {
		if c.IsSet(f) {
			switch f {
			case "region":
				a.AWSRegionIsSet = true
			case "domain":
				a.DomainIsSet = true
			case "tls-cert":
				a.TLSCertIsSet = true
			case "tls-key":
				a.TLSKeyIsSet = true
			case "workers":
				a.WorkerCountIsSet = true
			case "worker-size":
				a.WorkerSizeIsSet = true
			case "web-size":
				a.WebSizeIsSet = true
			case "iaas":
				a.IAASIsSet = true
			case "self-update":
				a.SelfUpdateIsSet = true
			case "db-size":
				a.DBSizeIsSet = true
			case "spot":
				a.SpotIsSet = true
			case "allow-ips":
				a.AllowIPsIsSet = true
			case "github-auth-client-id":
				a.GithubAuthClientIDIsSet = true
			case "github-auth-client-secret":
				a.GithubAuthClientSecretIsSet = true
			case "add-tag":
				a.TagsIsSet = true
			case "namespace":
				a.NamespaceIsSet = true
			case "zone":
				a.ZoneIsSet = true
			case "worker-type":
				a.WorkerTypeIsSet = true
			default:
				return fmt.Errorf("flag %q is not supported by deployment flags", f)
			}
		}
	}
	return nil
}

// WorkerSizes are the permitted concourse worker sizes
var WorkerSizes = []string{"medium", "large", "xlarge", "2xlarge", "4xlarge", "12xlarge", "24xlarge"}

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
func (a *DeployArgs) ModifyGithub(GithubAuthClientID, GithubAuthClientSecret string, GithubAuthIsSet bool) {
	a.GithubAuthClientID = GithubAuthClientID
	a.GithubAuthClientSecret = GithubAuthClientSecret
	a.GithubAuthIsSet = GithubAuthIsSet
}

// Validate validates that flag interdependencies
func (a DeployArgs) Validate() error {
	if err := a.validateCertFields(); err != nil {
		return err
	}

	if err := a.validateWorkerFields(); err != nil {
		return err
	}

	if err := a.validateWebFields(); err != nil {
		return err
	}

	if err := a.validateDBFields(); err != nil {
		return err
	}

	if err := a.validateGithubFields(); err != nil {
		return err
	}

	if err := a.validateTags(); err != nil {
		return err
	}

	return nil
}

func (a DeployArgs) validateCertFields() error {
	if a.TLSKey != "" && a.TLSCert == "" {
		return errors.New("--tls-key requires --tls-cert to also be provided")
	}
	if a.TLSCert != "" && a.TLSKey == "" {
		return errors.New("--tls-cert requires --tls-key to also be provided")
	}
	if (a.TLSKey != "" || a.TLSCert != "") && a.Domain == "" {
		return errors.New("custom certificates require --domain to be provided")
	}

	return nil
}

func (a DeployArgs) validateWorkerFields() error {
	if a.WorkerCount < 1 {
		return errors.New("minimum number of workers is 1")
	}

	for _, size := range WorkerSizes {
		if size == a.WorkerSize {
			return nil
		}
	}
	return fmt.Errorf("unknown worker size: `%s`. Valid sizes are: %v", a.WorkerSize, WorkerSizes)
}

func (a DeployArgs) validateWebFields() error {
	for _, size := range WebSizes {
		if size == a.WebSize {
			return nil
		}
	}
	return fmt.Errorf("unknown web node size: `%s`. Valid sizes are: %v", a.WebSize, WebSizes)
}

func (a DeployArgs) validateDBFields() error {
	if _, ok := DBSizes[a.DBSize]; !ok {
		return fmt.Errorf("unknown DB size: `%s`. Valid sizes are: %v", a.DBSize, DBSizes)
	}

	return nil
}

func (a DeployArgs) validateGithubFields() error {
	if a.GithubAuthClientID != "" && a.GithubAuthClientSecret == "" {
		return errors.New("--github-auth-client-id requires --github-auth-client-secret to also be provided")
	}
	if a.GithubAuthClientID == "" && a.GithubAuthClientSecret != "" {
		return errors.New("--github-auth-client-secret requires --github-auth-client-id to also be provided")
	}

	return nil
}

func (a DeployArgs) validateTags() error {
	for _, tag := range a.Tags {
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
