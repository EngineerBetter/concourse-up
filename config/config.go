package config

import (
	"fmt"
	"strings"

	"github.com/EngineerBetter/concourse-up/util"
)

// Config represents a concourse-up configuration file
type Config struct {
	AvailabilityZone          string `json:"availability_zone"`
	CredhubURL                string `json:"credhub_url"`
	CredhubUsername           string `json:"credhub_username"`
	CredhubPassword           string `json:"credhub_password"`
	CredhubAdminClientSecret  string `json:"credhub_admin_client_secret"`
	CredhubCACert             string `json:"credhub_ca_cert"`
	ConcourseCACert           string `json:"concourse_ca_cert"`
	ConcourseCert             string `json:"concourse_cert"`
	ConcourseDBName           string `json:"concourse_db_name"`
	ConcoursePassword         string `json:"concourse_password"`
	ConcourseUserProvidedCert bool   `json:"concourse_user_provided_cert"`
	ConcourseUsername         string `json:"concourse_username"`
	ConcourseWebSize          string `json:"concourse_web_size"`
	ConcourseWorkerCount      int    `json:"concourse_worker_count"`
	ConcourseWorkerSize       string `json:"concourse_worker_size"`
	ConfigBucket              string `json:"config_bucket"`
	Deployment                string `json:"deployment"`
	DirectorCACert            string `json:"director_ca_cert"`
	DirectorCert              string `json:"director_cert"`
	DirectorHMUserPassword    string `json:"director_hm_user_password"`
	DirectorKey               string `json:"director_key"`
	DirectorMbusPassword      string `json:"director_mbus_password"`
	DirectorNATSPassword      string `json:"director_nats_password"`
	DirectorPassword          string `json:"director_password"`
	DirectorPublicIP          string `json:"director_public_ip"`
	DirectorRegistryPassword  string `json:"director_registry_password"`
	DirectorUsername          string `json:"director_username"`
	Domain                    string `json:"domain"`
	EncryptionKey             string `json:"encryption_key"`
	GrafanaPassword           string `json:"grafana_password"`
	GrafanaUsername           string `json:"grafana_username"`
	HostedZoneID              string `json:"hosted_zone_id"`
	HostedZoneRecordPrefix    string `json:"hosted_zone_record_prefix"`
	MultiAZRDS                bool   `json:"multi_az_rds"`
	PrivateKey                string `json:"private_key"`
	Project                   string `json:"project"`
	PublicKey                 string `json:"public_key"`
	RDSDefaultDatabaseName    string `json:"rds_default_database_name"`
	RDSInstanceClass          string `json:"rds_instance_class"`
	RDSPassword               string `json:"rds_password"`
	RDSUsername               string `json:"rds_username"`
	Region                    string `json:"region"`
	SourceAccessIP            string `json:"source_access_ip"`
	TFStatePath               string `json:"tf_state_path"`
	AllowIPs                  string `json:"allow_ips"`
}

func generateDefaultConfig(iaas, project, deployment, configBucket, region string) (*Config, error) {
	privateKey, publicKey, _, err := util.GenerateSSHKeyPair()
	if err != nil {
		return nil, err
	}

	conf := Config{
		AvailabilityZone:         fmt.Sprintf("%sa", region),
		ConcourseDBName:          "concourse_atc",
		ConcourseWorkerCount:     1,
		ConcourseWebSize:         "small",
		ConcourseWorkerSize:      "xlarge",
		ConfigBucket:             configBucket,
		Deployment:               deployment,
		DirectorHMUserPassword:   util.GeneratePassword(),
		DirectorMbusPassword:     util.GeneratePassword(),
		DirectorNATSPassword:     util.GeneratePassword(),
		DirectorPassword:         util.GeneratePassword(),
		DirectorRegistryPassword: util.GeneratePassword(),
		DirectorUsername:         "admin",
		EncryptionKey:            util.GeneratePasswordWithLength(32),
		MultiAZRDS:               false,
		PrivateKey:               strings.TrimSpace(string(privateKey)),
		Project:                  project,
		PublicKey:                strings.TrimSpace(string(publicKey)),
		RDSDefaultDatabaseName:   "bosh",
		RDSPassword:              util.GeneratePassword(),
		RDSUsername:              "admin" + util.GeneratePassword(),
		Region:                   region,
		TFStatePath:              terraformStateFileName,
	}

	return &conf, nil
}

func updateConfig(c *Config, rdsInstanceClass string, ingressAddresses cidrBlocks) error {
	c.RDSInstanceClass = rdsInstanceClass
	c.AllowIPs = ingressAddresses.String()
	return nil
}
