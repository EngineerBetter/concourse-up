package config

import (
	"fmt"
	"strings"

	"github.com/EngineerBetter/concourse-up/util"
)

// Config represents a concourse-up configuration file
type Config struct {
	AvailabilityZone          string `json:"availability_zone"`
	ConcourseCACert           string `json:"concourse_ca_cert"`
	ConcourseCert             string `json:"concourse_cert"`
	ConcourseDBName           string `json:"concourse_db_name"`
	ConcourseKey              string `json:"concourse_key"`
	ConcoursePassword         string `json:"concourse_password"`
	ConcourseUsername         string `json:"concourse_username"`
	ConcourseUserProvidedCert bool   `json:"concourse_user_provided_cert"`
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
	InfluxDBPassword          string `json:"influxdb_password"`
	InfluxDBUsername          string `json:"influxdb_username"`
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
}

func generateDefaultConfig(iaas, project, deployment, configBucket, region, rdsInstanceClass string) (*Config, error) {
	privateKey, publicKey, err := util.GenerateSSHKeyPair()
	if err != nil {
		return nil, err
	}

	concourseUsername := "admin"
	concoursePassword := util.GeneratePassword()

	conf := Config{
		AvailabilityZone:         fmt.Sprintf("%sa", region),
		ConcourseDBName:          "concourse_atc",
		ConcoursePassword:        concoursePassword,
		ConcourseUsername:        concourseUsername,
		ConcourseWorkerCount:     1,
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
		GrafanaPassword:          concoursePassword,
		GrafanaUsername:          concourseUsername,
		InfluxDBPassword:         util.GeneratePassword(),
		InfluxDBUsername:         "admin",
		MultiAZRDS:               false,
		PrivateKey:               strings.TrimSpace(string(privateKey)),
		Project:                  project,
		PublicKey:                strings.TrimSpace(string(publicKey)),
		RDSDefaultDatabaseName:   "bosh",
		RDSInstanceClass:         rdsInstanceClass,
		RDSPassword:              util.GeneratePassword(),
		RDSUsername:              "admin" + util.GeneratePassword(),
		Region:                   region,
		SourceAccessIP:           "",
		TFStatePath:              terraformStateFileName,
	}

	return &conf, nil
}
