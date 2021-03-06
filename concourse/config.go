package concourse

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/asaskevich/govalidator"
)

func (client *Client) getInitialConfig() (config.Config, bool, error) {
	priorConfigExists, err := client.configClient.ConfigExists()
	if err != nil {
		return config.Config{}, false, fmt.Errorf("error determining if config already exists [%v]", err)
	}

	var isDomainUpdated bool
	var conf config.Config
	if priorConfigExists {
		if client.deployArgs.NetworkCIDRIsSet || client.deployArgs.PrivateCIDRIsSet || client.deployArgs.PublicCIDRIsSet {
			return config.Config{}, false, fmt.Errorf("custom CIDRs cannot be applied after intial deploy")
		}
		conf, err = client.configClient.Load()
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error loading existing config [%v]", err)
		}
		err = writeConfigLoadedSuccessMessage(client.stdout)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error writing config loaded success message [%v]", err)
		}

		conf, isDomainUpdated, err = populateConfigWithDefaultsOrProvidedArguments(conf, false, client.deployArgs, client.provider)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error merging new options with existing config: [%v]", err)
		}

	} else {
		conf, err = newConfig(client.configClient, client.deployArgs, client.provider, client.passwordGenerator, client.eightRandomLetters, client.sshGenerator)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error generating new config: [%v]", err)
		}

		err = client.configClient.Update(conf)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error persisting new config after setting values [%v]", err)
		}

		isDomainUpdated = true
	}

	return conf, isDomainUpdated, nil
}

func newConfig(configClient config.IClient, deployArgs *deploy.Args, provider iaas.Provider, passwordGenerator func(int) string, eightRandomLetters func() string, sshGenerator func() ([]byte, []byte, string, error)) (config.Config, error) {
	conf := configClient.NewConfig()
	conf, err := populateConfigWithDefaults(conf, provider, passwordGenerator, sshGenerator)
	if err != nil {
		return config.Config{}, fmt.Errorf("error generating default config: [%v]", err)
	}

	conf, _, err = populateConfigWithDefaultsOrProvidedArguments(conf, true, deployArgs, provider)
	if err != nil {
		return config.Config{}, fmt.Errorf("error generating default config: [%v]", err)
	}

	// Stuff from concourse.Deploy()
	switch provider.IAAS() {
	case iaas.AWS: // nolint
		conf.RDSDefaultDatabaseName = fmt.Sprintf("bosh_%s", eightRandomLetters())
	case iaas.GCP: // nolint
		conf.RDSDefaultDatabaseName = fmt.Sprintf("bosh-%s", eightRandomLetters())
	}

	// Why do we do this here?
	provider.WorkerType(conf.ConcourseWorkerSize)
	conf.AvailabilityZone = provider.Zone(deployArgs.Zone)
	// End stuff from concourse.Deploy()

	return conf, nil
}

func populateConfigWithDefaults(conf config.Config, provider iaas.Provider, passwordGenerator func(int) string, sshGenerator func() ([]byte, []byte, string, error)) (config.Config, error) {
	const defaultPasswordLength = 20

	privateKey, publicKey, _, err := sshGenerator()
	if err != nil {
		return config.Config{}, fmt.Errorf("error generating SSH keypair for new config: [%v]", err)
	}

	conf.AvailabilityZone = ""
	conf.ConcourseWebSize = "small"
	conf.ConcourseWorkerCount = 1
	conf.ConcourseWorkerSize = "xlarge"
	conf.DirectorHMUserPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorMbusPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorNATSPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorRegistryPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorUsername = "admin"
	conf.EncryptionKey = passwordGenerator(32)
	conf.PrivateKey = strings.TrimSpace(string(privateKey))
	conf.PublicKey = strings.TrimSpace(string(publicKey))
	conf.RDSPassword = passwordGenerator(defaultPasswordLength)
	conf.RDSUsername = "admin" + passwordGenerator(7)
	conf.Spot = true
	conf = populateConfigWithDefaultCIDRs(conf, provider)

	return conf, nil
}

func populateConfigWithDefaultsOrProvidedArguments(conf config.Config, newConfigCreated bool, deployArgs *deploy.Args, provider iaas.Provider) (config.Config, bool, error) {
	allow, err := parseAllowedIPsCIDRs(deployArgs.AllowIPs)
	if err != nil {
		return config.Config{}, false, err
	}

	conf, err = updateAllowedIPs(conf, allow)
	if err != nil {
		return config.Config{}, false, err
	}

	if newConfigCreated {
		conf.IAAS = deployArgs.IAAS
	}

	if deployArgs.ZoneIsSet {
		// This is a safeguard for a redeployment where zone does not belong to the region where the original deployment has happened
		if !newConfigCreated && deployArgs.Zone != conf.AvailabilityZone {
			return config.Config{}, false, fmt.Errorf("Existing deployment uses zone %s and cannot change to zone %s", conf.AvailabilityZone, deployArgs.Zone)
		}
		conf.AvailabilityZone = deployArgs.Zone
	}
	if newConfigCreated {
		conf.IAAS = deployArgs.IAAS
	}
	if newConfigCreated || deployArgs.WorkerCountIsSet {
		conf.ConcourseWorkerCount = deployArgs.WorkerCount
	}
	if newConfigCreated || deployArgs.WorkerSizeIsSet {
		conf.ConcourseWorkerSize = deployArgs.WorkerSize
	}
	if newConfigCreated || deployArgs.WebSizeIsSet {
		conf.ConcourseWebSize = deployArgs.WebSize
	}
	if newConfigCreated || deployArgs.DBSizeIsSet {
		conf.RDSInstanceClass = provider.DBType(deployArgs.DBSize)
	}
	if newConfigCreated || deployArgs.GithubAuthIsSet {
		conf.GithubClientID = deployArgs.GithubAuthClientID
		conf.GithubClientSecret = deployArgs.GithubAuthClientSecret
		conf.GithubAuthIsSet = deployArgs.GithubAuthIsSet
	}
	if newConfigCreated || deployArgs.TagsIsSet {
		conf.Tags = deployArgs.Tags
	}
	if newConfigCreated || deployArgs.SpotIsSet {
		conf.Spot = deployArgs.Spot && deployArgs.Preemptible
	}
	if newConfigCreated || deployArgs.WorkerTypeIsSet {
		conf.WorkerType = deployArgs.WorkerType
	}

	if newConfigCreated {
		if hasCIDRFlagsSet(deployArgs, provider) {
			conf = populateConfigWithDeployArgsCIDRs(conf, deployArgs, provider)
		}
	} else {
		// Existing config, these values are mandatory but did not exist in older versions
		if isMissingCIDRs(conf, provider) {
			conf = populateConfigWithDefaultCIDRs(conf, provider)
		}
	}

	var isDomainUpdated bool
	if newConfigCreated || deployArgs.DomainIsSet {
		if conf.Domain != deployArgs.Domain {
			isDomainUpdated = true
		}
		conf.Domain = deployArgs.Domain
	} else {
		if govalidator.IsIPv4(conf.Domain) {
			conf.Domain = ""
		}
	}

	return conf, isDomainUpdated, nil
}

func hasCIDRFlagsSet(deployArgs *deploy.Args, provider iaas.Provider) bool {
	switch provider.IAAS() {
	case iaas.AWS:
		return deployArgs.NetworkCIDRIsSet && deployArgs.PublicCIDRIsSet && deployArgs.PrivateCIDRIsSet
	case iaas.GCP:
		return deployArgs.PublicCIDRIsSet && deployArgs.PrivateCIDRIsSet
	default:
		return false
	}
}

func isMissingCIDRs(conf config.Config, provider iaas.Provider) bool {
	switch provider.IAAS() {
	case iaas.AWS:
		return conf.NetworkCIDR == "" || conf.PrivateCIDR == "" || conf.PublicCIDR == "" || conf.RDS1CIDR == "" || conf.RDS2CIDR == ""
	case iaas.GCP:
		return conf.PrivateCIDR == "" || conf.PublicCIDR == ""
	default:
		return false
	}
}

func populateConfigWithDeployArgsCIDRs(conf config.Config, deployArgs *deploy.Args, provider iaas.Provider) config.Config {
	switch provider.IAAS() {
	case iaas.AWS:
		conf.NetworkCIDR = deployArgs.NetworkCIDR
		conf.PublicCIDR = deployArgs.PublicCIDR
		conf.PrivateCIDR = deployArgs.PrivateCIDR
		conf.RDS1CIDR = deployArgs.RDS1CIDR
		conf.RDS2CIDR = deployArgs.RDS2CIDR
	case iaas.GCP:
		conf.PublicCIDR = deployArgs.PublicCIDR
		conf.PrivateCIDR = deployArgs.PrivateCIDR
	}
	return conf
}

func populateConfigWithDefaultCIDRs(conf config.Config, provider iaas.Provider) config.Config {
	switch provider.IAAS() {
	case iaas.AWS:
		conf.NetworkCIDR = "10.0.0.0/16"
		conf.PrivateCIDR = "10.0.1.0/24"
		conf.PublicCIDR = "10.0.0.0/24"
		conf.RDS1CIDR = "10.0.4.0/24"
		conf.RDS2CIDR = "10.0.5.0/24"
	case iaas.GCP:
		conf.PrivateCIDR = "10.0.1.0/24"
		conf.PublicCIDR = "10.0.0.0/24"
	}
	return conf
}

func updateAllowedIPs(c config.Config, ingressAddresses cidrBlocks) (config.Config, error) {
	addr, err := ingressAddresses.String()
	if err != nil {
		return c, err
	}
	c.AllowIPs = addr
	return c, nil
}

type cidrBlocks []*net.IPNet

func parseAllowedIPsCIDRs(s string) (cidrBlocks, error) {
	var x cidrBlocks
	for _, ip := range strings.Split(s, ",") {
		ip = strings.TrimSpace(ip)
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			ipNet = &net.IPNet{
				IP:   net.ParseIP(ip),
				Mask: net.CIDRMask(32, 32),
			}
		}
		if ipNet.IP == nil {
			return nil, fmt.Errorf("could not parse %q as an IP address or CIDR range", ip)
		}
		x = append(x, ipNet)
	}
	return x, nil
}

func (b cidrBlocks) String() (string, error) {
	var buf bytes.Buffer
	for i, ipNet := range b {
		if i > 0 {
			_, err := fmt.Fprintf(&buf, ", %q", ipNet)
			if err != nil {
				return "", err
			}
		} else {
			_, err := fmt.Fprintf(&buf, "%q", ipNet)
			if err != nil {
				return "", err
			}
		}
	}
	return buf.String(), nil
}
