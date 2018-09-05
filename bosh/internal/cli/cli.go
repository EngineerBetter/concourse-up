package cli

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ghodss/yaml"
)

type BOSHCLI struct {
	execCmd func(string, ...string) *exec.Cmd
}

type Option func(*BOSHCLI) error

func New(ops ...Option) (*BOSHCLI, error) {
	c := &BOSHCLI{
		execCmd: exec.Command,
	}
	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

type IAASConfig interface {
	Operations() []string
	Vars() map[string]interface{}
	Address() string
}

type Store interface {
	Set(key string, value []byte) error
	// Get must return a zero length byte slice and a nil error when the key is not present in the store
	Get(string) ([]byte, error)
}

type Environment struct {
	Address  string
	Username string
	Password string
	CACert   string
}

func (e Environment) env() []string {
	return append(os.Environ(),
		"BOSH_ENVIRONMENT=https://"+e.Address+":25555",
		"BOSH_CLIENT="+e.Username,
		"BOSH_CLIENT_SECRET="+e.Password,
		"BOSH_CA_CERT="+e.CACert,
	)
}

func (c *BOSHCLI) CreateEnv(store Store, config IAASConfig, password, cert, key, ca string) error {
	const stateFilename = "state.json"
	const varsFilename = "vars.yaml"
	vars := config.Vars()
	vars["admin_password"] = password
	vars["director_ssl.certificate"] = cert
	vars["director_ssl.private_key"] = key
	vars["director_ssl.ca"] = ca
	interpolatedManifest, err := Interpolate(directorManifest, config.Operations(), vars)
	if err != nil {
		return err
	}
	err = initJSON(store, stateFilename)
	if err != nil {
		return err
	}
	statePath, uploadState, err := writeToDisk(store, stateFilename)
	if err != nil {
		return err
	}
	defer uploadState()
	varsPath, uploadVars, err := writeToDisk(store, varsFilename)
	if err != nil {
		return err
	}
	defer uploadVars()
	manifestPath, err := writeTempFile([]byte(interpolatedManifest))
	if err != nil {
		return err
	}
	defer os.Remove(manifestPath)
	// cmd := c.execCmd("bosh", "create-env", "--state="+statePath, "--vars-store="+varsPath, manifestPath)
	cmd := c.execCmd("bosh", "create-env", "--state="+statePath, "--vars-store="+varsPath, manifestPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func readYAML(path string) (map[string]interface{}, error) {
	p, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var x map[string]interface{}
	err = yaml.Unmarshal(p, &x)
	return x, err
}

func getKey(m map[string]interface{}, path []string) (string, error) {
	v, ok := m[path[0]]
	if !ok {
		return "", errors.New("not found")
	}
	if len(path) == 1 {
		return v.(string), nil
	}
	return getKey(v.(map[string]interface{}), path[1:])
}

func initJSON(store Store, key string) error {
	data, err := store.Get(key)
	if err != nil {
		return err
	}
	if len(data) != 0 {
		return nil
	}
	return store.Set(key, []byte(`{}`))
}

func writeToDisk(store Store, key string) (filename string, upload func() error, err error) {
	data, err := store.Get(key)
	if err != nil {
		return "", nil, err
	}
	path, err := writeTempFile(data)
	if err != nil {
		return "", nil, err
	}
	upload = func() error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		os.Remove(path)
		return store.Set(key, data)
	}
	return path, upload, nil
}

func (c *BOSHCLI) UpdateCloudConfig(env Environment, cc string) error {
	ccPath, err := writeTempFile([]byte(cc))
	if err != nil {
		return err
	}
	defer os.Remove(ccPath)
	cmd := c.execCmd("bosh", "-n", "update-cloud-config", ccPath)
	cmd.Stderr = os.Stderr
	cmd.Env = env.env()
	return cmd.Run()
}

func (c *BOSHCLI) UploadStemcell(env Environment, ss string) error {
	cmd := c.execCmd("bosh", "-n", "upload-stemcell", ss)
	cmd.Stderr = os.Stderr
	cmd.Env = env.env()
	return cmd.Run()
}

func (c *BOSHCLI) Deploy(env Environment, name string, args ...string) error {
	a := []string{"-n", "--deployment", name, "deploy"}
	cmd := c.execCmd("bosh", append(a, args...)...)
	cmd.Stderr = os.Stderr
	cmd.Env = env.env()
	return cmd.Run()
}

func (c *BOSHCLI) UpdateDeployment(manifest string) error {
	return nil
}

func Interpolate(f string, ops []string, vars map[string]interface{}) (string, error) {
	args := []string{"interpolate"}
	for _, op := range ops {
		of, err := writeTempFile([]byte(op))
		if err != nil {
			return "", err
		}
		defer os.Remove(of)
		args = append(args, "-o="+of)
	}
	for k, v := range vars {
		encodedV, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		args = append(args, "-v="+k+"="+string(encodedV))
	}
	args = append(args, "-")
	cmd := exec.Command("bosh", args...)
	cmd.Stdin = strings.NewReader(f)
	var buf strings.Builder
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeTempFile(data []byte) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	name := f.Name()
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(name)
	}
	return name, err
}

const directorManifest = `---
name: bosh

releases:
- name: bosh
  version: "267.5.0"
  url: https://s3.amazonaws.com/bosh-compiled-release-tarballs/bosh-267.5.0-ubuntu-xenial-97.12-20180820-234355-048784588-20180820234358.tgz?versionId=X168wrj6izNJTi0V0bSPNpeKl_kZeahW
  sha1: 37499312e1186237434ee7a2492812b084c4e345
- name: bpm
  version: "0.11.0"
  url: https://s3.amazonaws.com/bosh-compiled-release-tarballs/bpm-0.11.0-ubuntu-xenial-97.12-20180820-235033-474094256-20180820235039.tgz?versionId=hzs0XSVxZzPFsTPDY3c3trwPf0OtugqB
  sha1: 6008b48192ac38ce6ac6ff23ea5029ac2b5bc2b9

resource_pools:
- name: vms
  network: default
  env:
    bosh:
      password: '*'
      mbus:
        cert: ((mbus_bootstrap_ssl))

disk_pools:
- name: disks
  disk_size: 65_536

networks:
- name: default
  type: manual
  subnets:
  - range: ((internal_cidr))
    gateway: ((internal_gw))
    static: [((internal_ip))]
    dns: [8.8.8.8]

instance_groups:
- name: bosh
  instances: 1
  jobs:
  - {name: bpm, release: bpm}
  - {name: nats, release: bosh}
  - {name: postgres-9.4, release: bosh}
  - {name: blobstore, release: bosh}
  - {name: director, release: bosh}
  - {name: health_monitor, release: bosh}
  resource_pool: vms
  persistent_disk_pool: disks
  networks:
  - name: default
    static_ips: [((internal_ip))]
  properties:
    agent:
      mbus: nats://nats:((nats_password))@((internal_ip)):4222
      env:
        bosh:
          blobstores:
          - provider: dav
            options:
              # todo switch to using https
              endpoint: http://((internal_ip)):25250
              user: agent
              password: ((blobstore_agent_password))
              tls:
                cert:
                  ca: ((blobstore_ca.certificate))
    nats:
      address: ((internal_ip))
      user: nats
      password: ((nats_password))
      tls:
        ca: ((nats_server_tls.ca))
        client_ca:
          certificate: ((nats_ca.certificate))
          private_key: ((nats_ca.private_key))
        server:
          certificate: ((nats_server_tls.certificate))
          private_key: ((nats_server_tls.private_key))
        director:
          certificate: ((nats_clients_director_tls.certificate))
          private_key: ((nats_clients_director_tls.private_key))
        health_monitor:
          certificate: ((nats_clients_health_monitor_tls.certificate))
          private_key: ((nats_clients_health_monitor_tls.private_key))
    postgres: &db
      listen_address: 127.0.0.1
      host: 127.0.0.1
      user: postgres
      password: ((postgres_password))
      database: bosh
      adapter: postgres
    blobstore:
      address: ((internal_ip))
      port: 25250
      provider: dav
      director:
        user: director
        password: ((blobstore_director_password))
      agent:
        user: agent
        password: ((blobstore_agent_password))
      tls:
        cert:
          ca: ((blobstore_ca.certificate))
          certificate: ((blobstore_server_tls.certificate))
          private_key: ((blobstore_server_tls.private_key))
    director:
      address: 127.0.0.1
      name: ((director_name))
      db: *db
      flush_arp: true
      enable_post_deploy: true
      generate_vm_passwords: true
      enable_dedicated_status_worker: true
      enable_nats_delivered_templates: true
      workers: 4
      local_dns:
        enabled: true
      events:
        record_events: true
      ssl:
        key: ((director_ssl.private_key))
        cert: ((director_ssl.certificate))
      user_management:
        provider: local
        local:
          users:
          - name: admin
            password: ((admin_password))
          - name: hm
            password: ((hm_password))
    hm:
      director_account:
        user: hm
        password: ((hm_password))
        ca_cert: ((director_ssl.ca))
      resurrector_enabled: true
    ntp: &ntp
    - time1.google.com
    - time2.google.com
    - time3.google.com
    - time4.google.com

cloud_provider:
  mbus: https://mbus:((mbus_bootstrap_password))@((internal_ip)):6868
  cert: ((mbus_bootstrap_ssl))
  properties:
    agent: {mbus: "https://mbus:((mbus_bootstrap_password))@0.0.0.0:6868"}
    blobstore: {provider: local, path: /var/vcap/micro_bosh/data/cache}
    ntp: *ntp

variables:
- name: admin_password
  type: password
- name: blobstore_director_password
  type: password
- name: blobstore_agent_password
  type: password
- name: hm_password
  type: password
- name: mbus_bootstrap_password
  type: password
- name: nats_password
  type: password
- name: postgres_password
  type: password

- name: default_ca
  type: certificate
  options:
    is_ca: true
    common_name: ca

- name: mbus_bootstrap_ssl
  type: certificate
  options:
    ca: default_ca
    common_name: ((internal_ip))
    alternative_names: [((internal_ip))]

- name: director_ssl
  type: certificate
  options:
    ca: default_ca
    common_name: ((internal_ip))
    alternative_names: [((internal_ip))]

- name: nats_ca
  type: certificate
  options:
    is_ca: true
    common_name: default.nats-ca.bosh-internal

- name: nats_server_tls
  type: certificate
  options:
    ca: nats_ca
    common_name: default.nats.bosh-internal
    alternative_names: [((internal_ip))]
    extended_key_usage:
    - server_auth

- name: nats_clients_director_tls
  type: certificate
  options:
    ca: nats_ca
    common_name: default.director.bosh-internal
    extended_key_usage:
    - client_auth

- name: nats_clients_health_monitor_tls
  type: certificate
  options:
    ca: nats_ca
    common_name: default.hm.bosh-internal
    extended_key_usage:
    - client_auth

- name: blobstore_ca
  type: certificate
  options:
    is_ca: true
    common_name: default.blobstore-ca.bosh-internal

- name: blobstore_server_tls
  type: certificate
  options:
    ca: blobstore_ca
    common_name: ((internal_ip))
    alternative_names: [((internal_ip))]`
