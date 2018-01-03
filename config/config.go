package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/EngineerBetter/concourse-up/terraform"
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
	TokenPrivateKey           string `json:"token_private_key"`
	TokenPublicKey            string `json:"token_public_key"`
	TSAFingerprint            string `json:"tsa_fingerprint"`
	TSAPrivateKey             string `json:"tsa_private_key"`
	TSAPublicKey              string `json:"tsa_public_key"`
	WorkerFingerprint         string `json:"worker_fingerprint"`
	WorkerPrivateKey          string `json:"worker_private_key"`
	WorkerPublicKey           string `json:"worker_public_key"`
	CredhubEncryptionPassword string `json:"credhub_encryption_password"`
	UAAAdmin                  string `json:"uaa_admin"`
	UAALogin                  string `json:"uaa_login"`
	UAAUsersAdmin             string `json:"uaa_users_admin"`
	UAAJWTPrivateKey          string `json:"uaa_jwt_private_key"`
	UAAJWTPublicKey           string `json:"uaa_jwt_public_key"`
}

func generateDefaultConfig(metadata *terraform.Metadata, iaas, project, deployment, configBucket, region, rdsInstanceClass string) (*Config, error) {
	privateKey, publicKey, _, err := util.GenerateSSHKeyPair()
	if err != nil {
		return nil, err
	}

	tsaPrivateKey, tsaPublicKey, tsaFingerprint, err := util.GenerateSSHKeyPair()
	if err != nil {
		return nil, err
	}

	workerPrivateKey, workerPublicKey, workerFingerprint, err := util.GenerateSSHKeyPair()
	if err != nil {
		return nil, err
	}

	tokenPrivateKey, tokenPublicKey, err := util.GenerateRSAKeyPair()
	if err != nil {
		return nil, err
	}

	uaaJWTPrivateKey, uaaJWTPublicKey, err := util.GenerateRSAKeyPair()
	if err != nil {
		return nil, err
	}

	caCert, caKey, err := util.GenerateCertificate(3*52*7*24*time.Hour, "acme co", nil, nil, nil)
	if err != nil {
		return nil, err
	}

	uaaCert, uaaKey, err := util.GenerateCertificate(3*52*7*24*time.Hour, "acme co", []string{"127.0.0.1", metadata.ATCPublicIP}, caCert, caKey)
	if err != nil {
		return nil, err
	}

	credhubCert, credhubKey, err := util.GenerateCertificate(3*52*7*24*time.Hour, "acme co", []string{"127.0.0.1", metadata.ATCPublicIP}, caCert, caKey)
	if err != nil {
		return nil, err
	}

	concourseUsername := "admin"
	concoursePassword := util.GeneratePassword()

	conf := Config{
		AvailabilityZone:          fmt.Sprintf("%sa", region),
		ConcourseDBName:           "concourse_atc",
		ConcoursePassword:         concoursePassword,
		ConcourseUsername:         concourseUsername,
		ConcourseWorkerCount:      1,
		ConcourseWebSize:          "small",
		ConcourseWorkerSize:       "xlarge",
		ConfigBucket:              configBucket,
		Deployment:                deployment,
		DirectorHMUserPassword:    util.GeneratePassword(),
		DirectorMbusPassword:      util.GeneratePassword(),
		DirectorNATSPassword:      util.GeneratePassword(),
		DirectorPassword:          util.GeneratePassword(),
		DirectorRegistryPassword:  util.GeneratePassword(),
		DirectorUsername:          "admin",
		EncryptionKey:             util.GeneratePasswordWithLength(32),
		GrafanaPassword:           concoursePassword,
		GrafanaUsername:           concourseUsername,
		InfluxDBPassword:          util.GeneratePassword(),
		InfluxDBUsername:          "admin",
		MultiAZRDS:                false,
		PrivateKey:                strings.TrimSpace(string(privateKey)),
		Project:                   project,
		PublicKey:                 strings.TrimSpace(string(publicKey)),
		RDSDefaultDatabaseName:    "bosh",
		RDSInstanceClass:          rdsInstanceClass,
		RDSPassword:               util.GeneratePassword(),
		RDSUsername:               "admin" + util.GeneratePassword(),
		Region:                    region,
		TFStatePath:               terraformStateFileName,
		TokenPrivateKey:           strings.TrimSpace(string(tokenPrivateKey)),
		TokenPublicKey:            strings.TrimSpace(string(tokenPublicKey)),
		TSAFingerprint:            strings.TrimSpace(tsaFingerprint),
		TSAPrivateKey:             strings.TrimSpace(string(tsaPrivateKey)),
		TSAPublicKey:              strings.TrimSpace(string(tsaPublicKey)),
		WorkerFingerprint:         strings.TrimSpace(workerFingerprint),
		WorkerPrivateKey:          strings.TrimSpace(string(workerPrivateKey)),
		WorkerPublicKey:           strings.TrimSpace(string(workerPublicKey)),
		CredhubEncryptionPassword: util.GeneratePasswordWithLength(40),
		UAAAdmin:                  util.GeneratePassword(),
		UAALogin:                  util.GeneratePassword(),
		UAAUsersAdmin:             util.GeneratePassword(),
		UAAJWTPrivateKey:          strings.TrimSpace(string(uaaJWTPrivateKey)),
		UAAJWTPublicKey:           strings.TrimSpace(string(uaaJWTPublicKey)),
	}

	return &conf, nil
}
