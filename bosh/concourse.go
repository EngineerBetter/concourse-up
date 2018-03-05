package bosh

import (
	"fmt"
	"io/ioutil"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/db"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

const concourseManifestFilename = "concourse.yml"
const credsFilename = "concourse-creds.yml"
const concourseDeploymentName = "concourse"

// ConcourseStemcellURL is a compile-time variable set with -ldflags
var ConcourseStemcellURL = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellURL"

// ConcourseStemcellVersion is a compile-time variable set with -ldflags
var ConcourseStemcellVersion = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion"

// ConcourseStemcellSHA1 is a compile-time variable set with -ldflags
var ConcourseStemcellSHA1 = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellSHA1"

// ConcourseReleaseURL is a compile-time variable set with -ldflags
var ConcourseReleaseURL = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseURL"

// ConcourseReleaseVersion is a compile-time variable set with -ldflags
var ConcourseReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseVersion"

// ConcourseReleaseSHA1 is a compile-time variable set with -ldflags
var ConcourseReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseSHA1"

// GardenReleaseURL is a compile-time variable set with -ldflags
var GardenReleaseURL = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseURL"

// GardenReleaseVersion is a compile-time variable set with -ldflags
var GardenReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseVersion"

// GardenReleaseSHA1 is a compile-time variable set with -ldflags
var GardenReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseSHA1"

// RiemannReleaseURL is a compile-time variable set with -ldflags
var RiemannReleaseURL = "COMPILE_TIME_VARIABLE_bosh_riemannReleaseURL"

// RiemannReleaseVersion is a compile-time variable set with -ldflags
var RiemannReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_riemannReleaseVersion"

// RiemannReleaseSHA1 is a compile-time variable set with -ldflags
var RiemannReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_riemannReleaseSHA1"

// GrafanaReleaseURL is a compile-time variable set with -ldflags
var GrafanaReleaseURL = "COMPILE_TIME_VARIABLE_bosh_grafanaReleaseURL"

// GrafanaReleaseVersion is a compile-time variable set with -ldflags
var GrafanaReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_grafanaReleaseVersion"

// GrafanaReleaseSHA1 is a compile-time variable set with -ldflags
var GrafanaReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_grafanaReleaseSHA1"

// InfluxDBReleaseURL is a compile-time variable set with -ldflags
var InfluxDBReleaseURL = "COMPILE_TIME_VARIABLE_bosh_influxDBReleaseURL"

// InfluxDBReleaseVersion is a compile-time variable set with -ldflags
var InfluxDBReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_influxDBReleaseVersion"

// InfluxDBReleaseSHA1 is a compile-time variable set with -ldflags
var InfluxDBReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_influxDBReleaseSHA1"

// CredhubReleaseURL is a compile-time variable set with -ldflags
var CredhubReleaseURL = "COMPILE_TIME_VARIABLE_bosh_CredhubReleaseURL"

// CredhubReleaseVersion is a compile-time variable set with -ldflags
var CredhubReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_CredhubReleaseVersion"

// CredhubReleaseSHA1 is a compile-time variable set with -ldflags
var CredhubReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_CredhubReleaseSHA1"

// UAAReleaseURL is a compile-time variable set with -ldflags
var UAAReleaseURL = "COMPILE_TIME_VARIABLE_bosh_UAAReleaseURL"

// UAAReleaseVersion is a compile-time variable set with -ldflags
var UAAReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_UAAReleaseVersion"

// UAAReleaseSHA1 is a compile-time variable set with -ldflags
var UAAReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_UAAReleaseSHA1"

func (client *Client) uploadConcourseStemcell() error {
	return client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		false,
		"upload-stemcell",
		ConcourseStemcellURL,
	)
}

func (client *Client) uploadConcourseReleases() error {
	for _, release := range []string{ConcourseReleaseURL, GardenReleaseURL, GrafanaReleaseURL, RiemannReleaseURL, InfluxDBReleaseURL, UAAReleaseURL, CredhubReleaseURL} {
		err := client.director.RunAuthenticatedCommand(
			client.stdout,
			client.stderr,
			false,
			"upload-release",
			"--stemcell",
			"ubuntu-trusty/"+ConcourseStemcellVersion,
			release,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) deployConcourse(creds []byte, detach bool) (newCreds []byte, err error) {
	concourseManifestBytes, err := generateConcourseManifest(client.config, client.metadata)
	if err != nil {
		return
	}

	concourseManifestPath, err := client.director.SaveFileToWorkingDir(concourseManifestFilename, concourseManifestBytes)
	if err != nil {
		return
	}

	credsPath, err := client.director.SaveFileToWorkingDir(credsFilename, creds)
	if err != nil {
		return
	}

	err = client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		detach,
		"--deployment",
		concourseDeploymentName,
		"deploy",
		concourseManifestPath,
		"--vars-store",
		credsPath,
	)
	newCreds, err1 := ioutil.ReadFile(credsPath)
	if err == nil {
		err = err1
	}
	return
}

func generateConcourseManifest(config *config.Config, metadata *terraform.Metadata) ([]byte, error) {
	templateParams := awsConcourseManifestParams{
		AllowSelfSignedCerts:    "true",
		ATCPublicIP:             metadata.ATCPublicIP.Value,
		ConcourseReleaseSHA1:    ConcourseReleaseSHA1,
		ConcourseReleaseVersion: ConcourseReleaseVersion,
		DBCACert:                db.RDSRootCert,
		DBHost:                  metadata.BoshDBAddress.Value,
		DBName:                  config.ConcourseDBName,
		DBPassword:              config.RDSPassword,
		DBPort:                  metadata.BoshDBPort.Value,
		DBUsername:              config.RDSUsername,
		EncryptionKey:           config.EncryptionKey,
		GardenReleaseSHA1:       GardenReleaseSHA1,
		GardenReleaseVersion:    GardenReleaseVersion,
		GrafanaPassword:         config.GrafanaPassword,
		GrafanaPort:             "3000",
		GrafanaReleaseSHA1:      GrafanaReleaseSHA1,
		GrafanaReleaseVersion:   GrafanaReleaseVersion,
		GrafanaURL:              fmt.Sprintf("https://%s:3000", config.Domain),
		GrafanaUsername:         config.GrafanaUsername,
		InfluxDBPassword:        config.InfluxDBPassword,
		InfluxDBReleaseSHA1:     InfluxDBReleaseSHA1,
		InfluxDBReleaseVersion:  InfluxDBReleaseVersion,
		InfluxDBUsername:        config.InfluxDBUsername,
		CredhubReleaseSHA1:      CredhubReleaseSHA1,
		CredhubReleaseVersion:   CredhubReleaseVersion,
		UAAReleaseSHA1:          UAAReleaseSHA1,
		UAAReleaseVersion:       UAAReleaseVersion,
		Password:                config.ConcoursePassword,
		Project:                 config.Project,
		RiemannReleaseSHA1:      RiemannReleaseSHA1,
		RiemannReleaseVersion:   RiemannReleaseVersion,
		StemcellSHA1:            ConcourseStemcellSHA1,
		StemcellURL:             ConcourseStemcellURL,
		StemcellVersion:         ConcourseStemcellVersion,
		TLSCert:                 config.ConcourseCert,
		TLSKey:                  config.ConcourseKey,
		TokenPrivateKey:         config.TokenPrivateKey,
		TokenPublicKey:          config.TokenPublicKey,
		TSAFingerprint:          config.TSAFingerprint,
		TSAPrivateKey:           config.TSAPrivateKey,
		TSAPublicKey:            config.TSAPublicKey,
		URL:                     fmt.Sprintf("https://%s", config.Domain),
		Username:                config.ConcourseUsername,
		WorkerCount:             config.ConcourseWorkerCount,
		WorkerSize:              config.ConcourseWorkerSize,
		WebSize:                 config.ConcourseWebSize,
		WorkerFingerprint:       config.WorkerFingerprint,
		WorkerPrivateKey:        config.WorkerPrivateKey,
		WorkerPublicKey:         config.WorkerPublicKey,
	}
	return util.RenderTemplate(awsConcourseManifestTemplate, templateParams)
}

type awsConcourseManifestParams struct {
	ATCPublicIP             string
	AllowSelfSignedCerts    string
	ConcourseReleaseSHA1    string
	ConcourseReleaseVersion string
	DBCACert                string
	DBHost                  string
	DBName                  string
	DBPassword              string
	DBPort                  string
	DBUsername              string
	EncryptionKey           string
	GardenReleaseSHA1       string
	GardenReleaseVersion    string
	GrafanaPassword         string
	GrafanaPort             string
	GrafanaReleaseSHA1      string
	GrafanaReleaseVersion   string
	GrafanaURL              string
	GrafanaUsername         string
	InfluxDBPassword        string
	InfluxDBReleaseSHA1     string
	InfluxDBReleaseVersion  string
	InfluxDBUsername        string
	CredhubReleaseSHA1      string
	CredhubReleaseVersion   string
	UAAReleaseSHA1          string
	UAAReleaseVersion       string
	Password                string
	Project                 string
	RiemannReleaseSHA1      string
	RiemannReleaseVersion   string
	StemcellSHA1            string
	StemcellURL             string
	StemcellVersion         string
	TLSCert                 string
	TLSKey                  string
	TokenPrivateKey         string
	TokenPublicKey          string
	TSAFingerprint          string
	TSAPrivateKey           string
	TSAPublicKey            string
	URL                     string
	Username                string
	WebSize                 string
	WorkerCount             int
	WorkerSize              string
	WorkerFingerprint       string
	WorkerPrivateKey        string
	WorkerPublicKey         string
}

// Indent is a helper function to indent the field a given number of spaces
func (params awsConcourseManifestParams) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

//go:generate go-bindata -pkg $GOPACKAGE assets/
var awsConcourseManifestTemplate = string(MustAsset("assets/concourse.yml"))
