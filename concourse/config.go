package concourse

import (
	"bytes"
	"fmt"
	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/util"
	"github.com/asaskevich/govalidator"
	"net"
	"strings"
)

func (client *Client) getInitialConfig() (config.Config, bool, error) {
	priorConfigExists, err := client.configClient.ConfigExists()
	if err != nil {
		return config.Config{}, false, fmt.Errorf("error determining if config already exists [%v]", err)
	}

	var isDomainUpdated bool
	var conf config.Config
	if priorConfigExists {
		conf, err = client.configClient.Load()
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error loading existing config [%v]", err)
		}
		writeConfigLoadedSuccessMessage(client.stdout)

		conf, isDomainUpdated, err = populateConfigWithDefaultsOrProvidedArguments(conf, false, client.deployArgs, client.provider)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error merging new options with existing config: [%v]", err)
		}

	} else {
		conf, err = newConfig(client.configClient, client.deployArgs, client.provider)
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

func newConfig(configClient config.IClient, deployArgs *deploy.Args, provider iaas.Provider) (config.Config, error) {
	conf := configClient.NewConfig()
	conf, err := populateConfigWithDefaults(conf)
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
		conf.RDSDefaultDatabaseName = fmt.Sprintf("bosh_%s", util.EightRandomLetters())
	case iaas.GCP: // nolint
		conf.RDSDefaultDatabaseName = fmt.Sprintf("bosh-%s", util.EightRandomLetters())
	}

	// Why do we do this here?
	provider.WorkerType(conf.ConcourseWorkerSize)
	conf.AvailabilityZone = provider.Zone(deployArgs.Zone)
	// End stuff from concourse.Deploy()

	return conf, nil
}

//RENAME ME
func populateConfigWithDefaults(conf config.Config) (config.Config, error) {
	privateKey, publicKey, _, err := util.GenerateSSHKeyPair()
	if err != nil {
		return config.Config{}, fmt.Errorf("error generating SSH keypair for new config: [%v]", err)
	}

	conf.AvailabilityZone = ""
	conf.ConcourseWorkerCount = 1
	conf.ConcourseWebSize = "small"
	conf.ConcourseWorkerSize = "xlarge"
	conf.DirectorHMUserPassword = util.GeneratePassword()
	conf.DirectorMbusPassword = util.GeneratePassword()
	conf.DirectorNATSPassword = util.GeneratePassword()
	conf.DirectorPassword = util.GeneratePassword()
	conf.DirectorRegistryPassword = util.GeneratePassword()
	conf.DirectorUsername = "admin"
	conf.EncryptionKey = util.GeneratePasswordWithLength(32)
	conf.PrivateKey = strings.TrimSpace(string(privateKey))
	conf.PublicKey = strings.TrimSpace(string(publicKey))
	conf.RDSPassword = util.GeneratePassword()
	conf.RDSUsername = "admin" + util.GeneratePasswordWithLength(7)
	conf.Spot = true
	conf.PrivateCIDR = "10.0.1.0/24"
	conf.PublicCIDR = "10.0.0.0/24"
	conf.NetworkCIDR = "10.0.0.0/16"
	conf.Rds1CIDR = "10.0.4.0/24"
	conf.Rds2CIDR = "10.0.5.0/24"

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
		if deployArgs.NetworkCIDRIsSet && deployArgs.PublicCIDRIsSet && deployArgs.PrivateCIDRIsSet {
			conf.NetworkCIDR = deployArgs.NetworkCIDR
			conf.PublicCIDR = deployArgs.PublicCIDR
			conf.PrivateCIDR = deployArgs.PrivateCIDR
			config.Rds1CIDR = deployArgs.Rds1CIDR
			config.Rds2CIDR = deployArgs.Rds2CIDR
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
