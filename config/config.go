package config

import (
	"encoding/json"
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
}

func generateDefaultConfig(project, deployment, configBucket, region string) ([]byte, error) {
	privateKey, publicKey, err := util.GenerateSSHKeyPair()
	if err != nil {
		return nil, err
	}

	conf := Config{
		ConcourseUsername:        "admin",
		ConcoursePassword:        util.GeneratePassword(),
		ConcourseWorkerCount:     1,
		ConcourseWorkerSize:      "xlarge",
		ConcourseDBName:          "concourse_atc",
		PublicKey:                strings.TrimSpace(string(publicKey)),
		PrivateKey:               strings.TrimSpace(string(privateKey)),
		Deployment:               deployment,
		ConfigBucket:             configBucket,
		RDSDefaultDatabaseName:   "bosh",
		SourceAccessIP:           "",
		Project:                  project,
		TFStatePath:              terraformStateFileName,
		Region:                   region,
		AvailabilityZone:         fmt.Sprintf("%sa", region),
		DirectorUsername:         "admin",
		DirectorPassword:         util.GeneratePassword(),
		DirectorHMUserPassword:   util.GeneratePassword(),
		EncryptionKey:            util.GeneratePasswordWithLength(32),
		DirectorMbusPassword:     util.GeneratePassword(),
		DirectorNATSPassword:     util.GeneratePassword(),
		DirectorRegistryPassword: util.GeneratePassword(),
		RDSInstanceClass:         "db.t2.small",
		RDSUsername:              "admin" + util.GeneratePassword(),
		RDSPassword:              util.GeneratePassword(),
		MultiAZRDS:               false,
	}

	return json.MarshalIndent(&conf, "", "  ")
}
