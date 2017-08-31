package bosh

import (
	"fmt"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/db"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

const concourseManifestFilename = "concourse.yml"
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

func (client *Client) uploadConcourseStemcell() error {
	return client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		false,
		"upload-stemcell",
		ConcourseStemcellURL,
	)
}

func (client *Client) deployConcourse(detach bool) error {
	concourseManifestBytes, err := generateConcourseManifest(client.config, client.metadata)
	if err != nil {
		return err
	}

	concourseManifestPath, err := client.director.SaveFileToWorkingDir(concourseManifestFilename, concourseManifestBytes)
	if err != nil {
		return err
	}

	if err = client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		detach,
		"--deployment",
		concourseDeploymentName,
		"deploy",
		concourseManifestPath,
	); err != nil {
		return err
	}

	return nil
}

func generateConcourseManifest(config *config.Config, metadata *terraform.Metadata) ([]byte, error) {
	templateParams := awsConcourseManifestParams{
		AllowSelfSignedCerts:    "true",
		ConcourseReleaseSHA1:    ConcourseReleaseSHA1,
		ConcourseReleaseURL:     ConcourseReleaseURL,
		ConcourseReleaseVersion: ConcourseReleaseVersion,
		DBCACert:                db.RDSRootCert,
		DBHost:                  metadata.BoshDBAddress.Value,
		DBName:                  config.ConcourseDBName,
		DBPassword:              config.RDSPassword,
		DBPort:                  metadata.BoshDBPort.Value,
		DBUsername:              config.RDSUsername,
		EncryptionKey:           config.EncryptionKey,
		GardenReleaseSHA1:       GardenReleaseSHA1,
		GardenReleaseURL:        GardenReleaseURL,
		GardenReleaseVersion:    GardenReleaseVersion,
		GrafanaPassword:         config.GrafanaPassword,
		GrafanaPort:             "3000",
		GrafanaReleaseSHA1:      GrafanaReleaseSHA1,
		GrafanaReleaseURL:       GrafanaReleaseURL,
		GrafanaReleaseVersion:   GrafanaReleaseVersion,
		GrafanaURL:              fmt.Sprintf("https://%s:3000", config.Domain),
		GrafanaUsername:         config.GrafanaUsername,
		InfluxDBPassword:        config.InfluxDBPassword,
		InfluxDBReleaseSHA1:     InfluxDBReleaseSHA1,
		InfluxDBReleaseURL:      InfluxDBReleaseURL,
		InfluxDBReleaseVersion:  InfluxDBReleaseVersion,
		InfluxDBUsername:        config.InfluxDBUsername,
		Password:                config.ConcoursePassword,
		Project:                 config.Project,
		RiemannReleaseSHA1:      RiemannReleaseSHA1,
		RiemannReleaseURL:       RiemannReleaseURL,
		RiemannReleaseVersion:   RiemannReleaseVersion,
		StemcellSHA1:            ConcourseStemcellSHA1,
		StemcellURL:             ConcourseStemcellURL,
		StemcellVersion:         ConcourseStemcellVersion,
		TLSCert:                 config.ConcourseCert,
		TLSKey:                  config.ConcourseKey,
		URL:                     fmt.Sprintf("https://%s", config.Domain),
		Username:                config.ConcourseUsername,
		WorkerCount:             config.ConcourseWorkerCount,
		WorkerSize:              config.ConcourseWorkerSize,
	}

	return util.RenderTemplate(awsConcourseManifestTemplate, templateParams)
}

type awsConcourseManifestParams struct {
	AllowSelfSignedCerts    string
	ConcourseReleaseSHA1    string
	ConcourseReleaseURL     string
	ConcourseReleaseVersion string
	DBCACert                string
	DBHost                  string
	DBName                  string
	DBPassword              string
	DBPort                  string
	DBUsername              string
	EncryptionKey           string
	GardenReleaseSHA1       string
	GardenReleaseURL        string
	GardenReleaseVersion    string
	GrafanaPassword         string
	GrafanaPort             string
	GrafanaReleaseSHA1      string
	GrafanaReleaseURL       string
	GrafanaReleaseVersion   string
	GrafanaURL              string
	GrafanaUsername         string
	InfluxDBPassword        string
	InfluxDBReleaseSHA1     string
	InfluxDBReleaseURL      string
	InfluxDBReleaseVersion  string
	InfluxDBUsername        string
	Password                string
	Project                 string
	RiemannReleaseSHA1      string
	RiemannReleaseURL       string
	RiemannReleaseVersion   string
	StemcellSHA1            string
	StemcellURL             string
	StemcellVersion         string
	TLSCert                 string
	TLSKey                  string
	URL                     string
	Username                string
	WorkerCount             int
	WorkerSize              string
}

// Indent is a helper function to indent the field a given number of spaces
func (params awsConcourseManifestParams) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

const awsConcourseManifestTemplate = `---
name: concourse

releases:
- name: concourse
  url: "<% .ConcourseReleaseURL %>"
  sha1: "<% .ConcourseReleaseSHA1 %>"
  version: <% .ConcourseReleaseVersion %>

- name: garden-runc
  url: "<% .GardenReleaseURL %>"
  sha1: "<% .GardenReleaseSHA1 %>"
  version: <% .GardenReleaseVersion %>

- name: riemann
  url: "<% .RiemannReleaseURL %>"
  sha1: "<% .RiemannReleaseSHA1 %>"
  version: <% .RiemannReleaseVersion %>

- name: grafana
  url: "<% .GrafanaReleaseURL %>"
  sha1: "<% .GrafanaReleaseSHA1 %>"
  version: <% .GrafanaReleaseVersion %>

- name: influxdb
  url: "<% .InfluxDBReleaseURL %>"
  sha1: "<% .InfluxDBReleaseSHA1 %>"
  version: <% .InfluxDBReleaseVersion %>

stemcells:
- alias: trusty
  os: ubuntu-trusty
  version: "<% .StemcellVersion %>"

tags:
  concourse-up-project: <% .Project %>
  concourse-up-component: concourse

instance_groups:
- name: web
  instances: 1
  vm_type: concourse-web
  stemcell: trusty
  azs:
  - z1
  networks:
  - name: public
    default: [dns, gateway]
    static_ips: [10.0.0.7]
  vm_extensions:
  - elb
  jobs:
  - name: atc
    release: concourse
    properties:
      allow_self_signed_certificates: <% .AllowSelfSignedCerts %>
      external_url: <% .URL %>
      encryption_key: <% .EncryptionKey %>
      basic_auth_username: <% .Username %>
      basic_auth_password: <% .Password %>
      tls_cert: |-
        <% .Indent "8" .TLSCert %>
      tls_key: |-
        <% .Indent "8" .TLSKey %>
      riemann:
        host: 127.0.0.1
        port: 5555

      postgresql:
        port: <% .DBPort %>
        database: <% .DBName %>
        role:
          name: <% .DBUsername %>
          password:  <% .DBPassword %>
        host: <% .DBHost %>
        ssl_mode: verify-full
        ca_cert: |-
          <% .Indent "10" .DBCACert %>

  - name: tsa
    release: concourse
    properties: {}
  - name: riemann
    release: riemann
    properties:
      riemann:
        influxdb:
          host: 127.0.0.1
          port: 8086
          password: <% .InfluxDBUsername %>
          username: <% .InfluxDBPassword %>
          database: riemann
  - name: influxdb
    release: influxdb
    properties:
      influxdb:
        database: riemann
        user: <% .InfluxDBUsername %>
        password: <% .InfluxDBPassword %>
  - name: riemann-emitter
    release: riemann
    properties:
      riemann_emitter:
        host: 127.0.0.1
        port: 5555
  - name: grafana
    release: grafana
    properties:
      grafana:
        admin_username: <% .GrafanaUsername %>
        admin_password: <% .GrafanaPassword %>
        listen_port: <% .GrafanaPort %>
        root_url: <% .GrafanaURL %>
        datasource:
          name: influxdb
          url: http://127.0.0.1:8086
          database_type: influxdb
          user: <% .InfluxDBUsername %>
          password: <% .InfluxDBPassword %>
          database_name: riemann
        ssl:
          cert: |-
            <% .Indent "12" .TLSCert %>
          key: |-
            <% .Indent "12" .TLSKey %>
        dashboards:
          - name: Concourse
            content: '{"__inputs":[],"__requires":[{"type":"grafana","id":"grafana","name":"Grafana","version":"4.4.1"},{"type":"panel","id":"graph","name":"Graph","version":""},{"type":"datasource","id":"influxdb","name":"InfluxDB","version":"1.0.0"}],"annotations":{"list":[]},"editable":true,"gnetId":null,"graphTooltip":0,"hideControls":false,"id":null,"links":[],"refresh":"10s","rows":[{"collapse":false,"height":250,"panels":[{"aliasColors":{},"bars":false,"dashLength":10,"dashes":false,"datasource":"influxdb","editable":true,"error":false,"fill":1,"grid":{},"height":"","id":4,"legend":{"avg":false,"current":false,"max":false,"min":false,"show":false,"total":false,"values":false},"lines":true,"linewidth":2,"links":[],"nullPointMode":"connected","percentage":false,"pointradius":5,"points":false,"renderer":"flot","seriesOverrides":[],"spaceLength":10,"span":6,"stack":false,"steppedLine":false,"targets":[{"dsType":"influxdb","groupBy":[{"params":["$interval"],"type":"time"},{"params":["host"],"type":"tag"},{"params":["null"],"type":"fill"}],"measurement":"cpu","orderByTime":"ASC","policy":"default","refId":"A","resultFormat":"time_series","select":[[{"params":["value"],"type":"field"},{"params":[],"type":"mean"}]],"tags":[]}],"thresholds":[],"timeFrom":null,"timeShift":null,"title":"CPU","tooltip":{"msResolution":false,"shared":true,"sort":0,"value_type":"cumulative"},"type":"graph","xaxis":{"buckets":null,"mode":"time","name":null,"show":true,"values":[]},"yaxes":[{"format":"percentunit","label":null,"logBase":1,"max":1,"min":0,"show":true},{"format":"short","label":null,"logBase":1,"max":null,"min":null,"show":true}]},{"aliasColors":{},"bars":false,"dashLength":10,"dashes":false,"datasource":"influxdb","editable":true,"error":false,"fill":1,"grid":{},"id":2,"legend":{"alignAsTable":false,"avg":false,"current":false,"max":false,"min":false,"rightSide":false,"show":false,"total":false,"values":false},"lines":true,"linewidth":2,"links":[],"nullPointMode":"connected","percentage":false,"pointradius":5,"points":false,"renderer":"flot","seriesOverrides":[],"spaceLength":10,"span":6,"stack":false,"steppedLine":false,"targets":[{"dsType":"influxdb","groupBy":[{"params":["$interval"],"type":"time"},{"params":["host"],"type":"tag"},{"params":["null"],"type":"fill"}],"measurement":"disk /var/vcap/data","orderByTime":"ASC","policy":"default","query":"SELECT mean(\"value\") FROM \"disk /var/vcap/data\" WHERE $timeFilter GROUP BY time($interval), \"host\" fill(null)","rawQuery":true,"refId":"A","resultFormat":"time_series","select":[[{"params":["value"],"type":"field"},{"params":[],"type":"mean"}]],"tags":[]}],"thresholds":[],"timeFrom":null,"timeShift":null,"title":"Disk Usage","tooltip":{"msResolution":false,"shared":true,"sort":0,"value_type":"cumulative"},"type":"graph","xaxis":{"buckets":null,"mode":"time","name":null,"show":true,"values":[]},"yaxes":[{"format":"percentunit","label":null,"logBase":1,"max":1,"min":0,"show":true},{"format":"short","label":null,"logBase":1,"max":null,"min":null,"show":true}]}],"repeat":null,"repeatIteration":null,"repeatRowId":null,"showTitle":false,"title":"Row","titleSize":"h6"},{"collapse":false,"height":"250px","panels":[{"aliasColors":{},"bars":false,"dashLength":10,"dashes":false,"datasource":"influxdb","editable":true,"error":false,"fill":0,"grid":{},"height":"","id":1,"legend":{"avg":false,"current":false,"max":false,"min":false,"show":false,"total":false,"values":false},"lines":true,"linewidth":2,"links":[],"nullPointMode":"connected","percentage":false,"pointradius":3,"points":true,"renderer":"flot","seriesOverrides":[],"spaceLength":10,"span":6,"stack":false,"steppedLine":false,"targets":[{"dsType":"influxdb","fields":[{"func":"mean","name":"value"}],"groupBy":[{"params":["$interval"],"type":"time"},{"params":["job"],"type":"tag"}],"groupByTags":["job"],"measurement":"build finished","orderByTime":"ASC","policy":"default","query":"SELECT mean(value) FROM \"build finished\" WHERE $timeFilter GROUP BY time($interval), \"job\"","refId":"A","resultFormat":"time_series","select":[[{"params":["value"],"type":"field"},{"params":[],"type":"mean"}]],"tags":[]}],"thresholds":[],"timeFrom":null,"timeShift":null,"title":"Build Durations","tooltip":{"msResolution":false,"shared":false,"sort":0,"value_type":"cumulative"},"type":"graph","xaxis":{"buckets":null,"mode":"time","name":null,"show":true,"values":[]},"yaxes":[{"format":"ms","logBase":1,"max":null,"min":null,"show":true},{"format":"short","logBase":1,"max":null,"min":null,"show":true}]},{"aliasColors":{},"bars":false,"dashLength":10,"dashes":false,"datasource":"influxdb","fill":1,"id":5,"legend":{"avg":false,"current":false,"max":false,"min":false,"show":true,"total":false,"values":false},"lines":true,"linewidth":1,"links":[],"nullPointMode":"null","percentage":false,"pointradius":5,"points":false,"renderer":"flot","seriesOverrides":[],"spaceLength":10,"span":6,"stack":false,"steppedLine":false,"targets":[{"dsType":"influxdb","groupBy":[{"params":["worker"],"type":"tag"}],"measurement":"worker containers","orderByTime":"ASC","policy":"default","refId":"A","resultFormat":"time_series","select":[[{"params":["value"],"type":"field"}]],"tags":[]}],"thresholds":[],"timeFrom":null,"timeShift":null,"title":"Containers","tooltip":{"shared":true,"sort":0,"value_type":"individual"},"type":"graph","xaxis":{"buckets":null,"mode":"time","name":null,"show":true,"values":[]},"yaxes":[{"format":"short","label":null,"logBase":1,"max":null,"min":null,"show":true},{"format":"short","label":null,"logBase":1,"max":null,"min":null,"show":true}]}],"repeat":null,"repeatIteration":null,"repeatRowId":null,"showTitle":false,"title":"New row","titleSize":"h6"},{"collapse":false,"height":250,"panels":[],"repeat":null,"repeatIteration":null,"repeatRowId":null,"showTitle":false,"title":"Dashboard Row","titleSize":"h6"}],"schemaVersion":14,"style":"dark","tags":[],"templating":{"list":[]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["5s","10s","30s","1m","5m","15m","30m","1h","2h","1d"],"time_options":["5m","15m","1h","6h","12h","24h","2d","7d","30d"]},"timezone":"browser","title":"Concourse","version":0}'

- name: worker
  instances: <% .WorkerCount %>
  vm_type: concourse-<% .WorkerSize %>
  stemcell: trusty
  azs:
  - z1
  networks:
  - name: private
    default: [dns, gateway]
  jobs:
  - name: groundcrew
    release: concourse
    properties: {}
  - name: baggageclaim
    release: concourse
    properties: {}
  - name: garden
    release: garden-runc
    properties:
      garden:
        listen_network: tcp
        listen_address: 0.0.0.0:7777
  - name: riemann-emitter
    release: riemann
    properties:
      riemann_emitter:
        host: 10.0.0.7
        port: 5555

update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000`
