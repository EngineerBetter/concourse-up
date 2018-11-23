package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/asaskevich/govalidator"
)

const terraformStateFileName = "terraform.tfstate"
const configFilePath = "config.json"

// IClient is an interface for the config file client
type IClient interface {
	Load() (Config, error)
	DeleteAll(config Config) error
	LoadOrCreate(deployArgs *DeployArgs) (Config, bool, bool, error)
	Update(Config) error
	StoreAsset(filename string, contents []byte) error
	HasAsset(filename string) (bool, error)
	LoadAsset(filename string) ([]byte, error)
	DeleteAsset(filename string) error
}

// Client is a client for loading the config file  from S3
type Client struct {
	Iaas         iaas.Provider
	Project      string
	Namespace    string
	BucketName   string
	BucketExists bool
	BucketError  error
	Config       *Config
}

// New instantiates a new client
func New(iaas iaas.Provider, project, namespace string) *Client {
	namespace = determineNamespace(namespace, iaas.Region())
	bucketName, exists, err := determineBucketName(iaas, namespace, project)

	if !exists && err == nil {
		err = iaas.CreateBucket(bucketName)
	}

	return &Client{
		iaas,
		project,
		namespace,
		bucketName,
		exists,
		err,
		&Config{},
	}
}

// StoreAsset stores an associated configuration file
func (client *Client) StoreAsset(filename string, contents []byte) error {
	return client.Iaas.WriteFile(client.configBucket(),
		filename,
		contents,
	)
}

// LoadAsset loads an associated configuration file
func (client *Client) LoadAsset(filename string) ([]byte, error) {
	return client.Iaas.LoadFile(
		client.configBucket(),
		filename,
	)
}

// DeleteAsset deletes an associated configuration file
func (client *Client) DeleteAsset(filename string) error {
	return client.Iaas.DeleteFile(
		client.configBucket(),
		filename,
	)
}

// HasAsset returns true if an associated configuration file exists
func (client *Client) HasAsset(filename string) (bool, error) {
	return client.Iaas.HasFile(
		client.configBucket(),
		filename,
	)
}

// Update stores the conconcourse up config file to S3
func (client *Client) Update(config Config) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return client.Iaas.WriteFile(client.configBucket(), configFilePath, bytes)
}

// DeleteAll deletes the entire configuration bucket
func (client *Client) DeleteAll(config Config) error {
	return client.Iaas.DeleteVersionedBucket(config.ConfigBucket)
}

// Load loads an existing config file from S3
func (client *Client) Load() (Config, error) {
	if client.BucketError != nil {
		return Config{}, client.BucketError
	}

	configBytes, err := client.Iaas.LoadFile(
		client.configBucket(),
		configFilePath,
	)
	if err != nil {
		return Config{}, err
	}

	conf := Config{}
	if err := json.Unmarshal(configBytes, &conf); err != nil {
		return Config{}, err
	}

	return conf, nil
}

// LoadOrCreate loads an existing config file from S3, or creates a default if one doesn't already exist
func (client *Client) LoadOrCreate(deployArgs *DeployArgs) (Config, bool, bool, error) {

	var isDomainUpdated bool

	if client.BucketError != nil {
		return Config{}, false, false, client.BucketError
	}

	config, err := generateDefaultConfig(
		client.Project,
		deployment(client.Project),
		client.configBucket(),
		client.Iaas.Region(),
		client.Namespace,
	)
	if err != nil {
		return Config{}, false, false, err
	}

	defaultConfigBytes, err := json.Marshal(&config)
	if err != nil {
		return Config{}, false, false, err
	}

	configBytes, newConfigCreated, err := client.Iaas.EnsureFileExists(
		client.configBucket(),
		configFilePath,
		defaultConfigBytes,
	)
	if err != nil {
		return Config{}, newConfigCreated, false, err
	}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, newConfigCreated, false, err
	}

	allow, err := parseCIDRBlocks(deployArgs.AllowIPs)
	if err != nil {
		return Config{}, newConfigCreated, false, err
	}

	config, err = updateAllowedIPs(config, allow)
	if err != nil {
		return Config{}, newConfigCreated, false, err
	}

	if newConfigCreated {
		config.IAAS = deployArgs.IAAS
	}

	if deployArgs.ZoneIsSet {
		// This is a safeguard for a redeployment where zone does not belong to the region where the original deployment has happened
		if !newConfigCreated && !strings.Contains(deployArgs.Zone, config.AvailabilityZone) {
			return Config{}, false, false, fmt.Errorf("Zone %s does not belong to region %s", deployArgs.Zone, config.AvailabilityZone)
		}
		config.AvailabilityZone = deployArgs.Zone
	}
	if newConfigCreated || deployArgs.WorkerCountIsSet {
		config.ConcourseWorkerCount = deployArgs.WorkerCount
	}
	if newConfigCreated || deployArgs.WorkerSizeIsSet {
		config.ConcourseWorkerSize = deployArgs.WorkerSize
	}
	if newConfigCreated || deployArgs.WebSizeIsSet {
		config.ConcourseWebSize = deployArgs.WebSize
	}
	if newConfigCreated || deployArgs.DBSizeIsSet {
		config.RDSInstanceClass = DBSizes[deployArgs.DBSize]
	}
	if newConfigCreated || deployArgs.GithubAuthIsSet {
		config.GithubClientID = deployArgs.GithubAuthClientID
		config.GithubClientSecret = deployArgs.GithubAuthClientSecret
		config.GithubAuthIsSet = deployArgs.GithubAuthIsSet
	}
	if newConfigCreated || deployArgs.TagsIsSet {
		config.Tags = deployArgs.Tags
	}
	if newConfigCreated || deployArgs.SpotIsSet {
		config.Spot = deployArgs.Spot
	}
	if newConfigCreated || deployArgs.WorkerTypeIsSet {
		config.WorkerType = deployArgs.WorkerType
	}

	if newConfigCreated || deployArgs.DomainIsSet {
		if config.Domain != deployArgs.Domain {
			isDomainUpdated = true
		}
		config.Domain = deployArgs.Domain
	} else {
		if govalidator.IsIPv4(config.Domain) {
			config.Domain = ""
		}
	}
	return config, newConfigCreated, isDomainUpdated, nil
}

func (client *Client) configBucket() string {
	return client.BucketName
}

func deployment(project string) string {
	return fmt.Sprintf("concourse-up-%s", project)
}

func createBucketName(deployment, extension string) string {
	return fmt.Sprintf("%s-%s-config", deployment, extension)
}

func determineBucketName(iaas iaas.Provider, namespace, project string) (string, bool, error) {
	regionBucketName := createBucketName(deployment(project), iaas.Region())
	namespaceBucketName := createBucketName(deployment(project), namespace)

	foundRegionNamedBucket, err := iaas.BucketExists(regionBucketName)
	if err != nil {
		return "", false, err
	}
	foundNamespacedBucket, err := iaas.BucketExists(namespaceBucketName)
	if err != nil {
		return "", false, err
	}

	foundOne := foundRegionNamedBucket || foundNamespacedBucket

	switch {
	case !foundRegionNamedBucket && foundNamespacedBucket:
		return namespaceBucketName, foundOne, nil
	case foundRegionNamedBucket && !foundNamespacedBucket:
		return regionBucketName, foundOne, nil
	default:
		return namespaceBucketName, foundOne, nil
	}
}

type cidrBlocks []*net.IPNet

func parseCIDRBlocks(s string) (cidrBlocks, error) {
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

func determineNamespace(namespace, region string) string {
	if namespace == "" {
		return region
	}
	return namespace
}
