package testsupport

import (
	"github.com/xenolf/lego/lego"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
)

// FakeProvider is a fake Provider
type FakeProvider struct {
	FakeAttr                          func(attr string) (string, error)
	FakeBucketExists                  func(name string) (bool, error)
	FakeCheckForWhitelistedIP         func(ip, securityGroup string) (bool, error)
	FakeCreateBucket                  func(name string) error
	FakeCreateDatabases               func(name, username, password string) error
	FakeDeleteFile                    func(bucket, path string) error
	FakeDeleteVersionedBucket         func(name string) error
	FakeDeleteVMsInDeployment         func(zone, project, deployment string) error
	FakeDeleteVMsInVPC                func(vpcID string) ([]string, error)
	FakeDeleteVolumes                 func(volumesToDelete []string, deleteVolume func(ec2Client iaas.IEC2, volumeID *string) error) error
	FakeEnsureFileExists              func(bucket, path string, defaultContents []byte) ([]byte, bool, error)
	FakeFindLongestMatchingHostedZone func(subdomain string) (string, string, error)
	FakeHasFile                       func(bucket, path string) (bool, error)
	FakeIAAS                          func() iaas.Name
	FakeLoadFile                      func(bucket, path string) ([]byte, error)
	FakeRegion                        func() string
	FakeWorkerType                    func(string)
	FakeWriteFile                     func(bucket, path string, contents []byte) error
	FakeZone                          func(string) string
	FakeDBType                        func(string) string
	FakeChoose                        func(iaas.Choice) interface{}
}

// Attr is a fake attr
func (fakeProvider *FakeProvider) Attr(attr string) (string, error) {
	return fakeProvider.FakeAttr(attr)
}

// BucketExists is a fake bucketexists
func (fakeProvider *FakeProvider) BucketExists(name string) (bool, error) {
	return fakeProvider.FakeBucketExists(name)
}

// CheckForWhitelistedIP is a fake checkforwhitelistedip
func (fakeProvider *FakeProvider) CheckForWhitelistedIP(ip, securityGroup string) (bool, error) {
	return fakeProvider.FakeCheckForWhitelistedIP(ip, securityGroup)
}

// CreateBucket is a fake createbucket
func (fakeProvider *FakeProvider) CreateBucket(name string) error {
	return fakeProvider.FakeCreateBucket(name)
}

// CreateDatabases is a fake createdatabases
func (fakeProvider *FakeProvider) CreateDatabases(name, username, password string) error {
	return fakeProvider.FakeCreateDatabases(name, username, password)
}

// DeleteFile is a fake deletefile
func (fakeProvider *FakeProvider) DeleteFile(bucket, path string) error {
	return fakeProvider.FakeDeleteFile(bucket, path)
}

// DeleteVersionedBucket is a fake deleteversionedbucket
func (fakeProvider *FakeProvider) DeleteVersionedBucket(name string) error {
	return fakeProvider.FakeDeleteVersionedBucket(name)
}

// DeleteVMsInDeployment is a fake deletevmsindeployment
func (fakeProvider *FakeProvider) DeleteVMsInDeployment(zone, project, deployment string) error {
	return fakeProvider.FakeDeleteVMsInDeployment(zone, project, deployment)
}

// DeleteVMsInVPC is a fake deletevmsinvpc
func (fakeProvider *FakeProvider) DeleteVMsInVPC(vpcID string) ([]string, error) {
	return fakeProvider.FakeDeleteVMsInVPC(vpcID)
}

// DeleteVolumes is a fake deletevolumes
func (fakeProvider *FakeProvider) DeleteVolumes(volumesToDelete []string, deleteVolume func(ec2Client iaas.IEC2, volumeID *string) error) error {
	return fakeProvider.FakeDeleteVolumes(volumesToDelete, deleteVolume)
}

// EnsureFileExists is a fake ensurefileexists
func (fakeProvider *FakeProvider) EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
	return fakeProvider.FakeEnsureFileExists(bucket, path, defaultContents)
}

// FindLongestMatchingHostedZone is a fake findlongestmatchinghostedzone
func (fakeProvider *FakeProvider) FindLongestMatchingHostedZone(subdomain string) (string, string, error) {
	return fakeProvider.FakeFindLongestMatchingHostedZone(subdomain)
}

// HasFile is a fake hasfile
func (fakeProvider *FakeProvider) HasFile(bucket, path string) (bool, error) {
	return fakeProvider.FakeHasFile(bucket, path)
}

// IAAS is a fake iaas
func (fakeProvider *FakeProvider) IAAS() iaas.Name {
	return fakeProvider.FakeIAAS()
}

// LoadFile is a fake loadfile
func (fakeProvider *FakeProvider) LoadFile(bucket, path string) ([]byte, error) {
	return fakeProvider.FakeLoadFile(bucket, path)
}

// Region is a fake region
func (fakeProvider *FakeProvider) Region() string {
	return fakeProvider.FakeRegion()
}

// WorkerType is a fake workertype
func (fakeProvider *FakeProvider) WorkerType(workerType string) {
	fakeProvider.FakeWorkerType(workerType)
}

// WriteFile is a fake writefile
func (fakeProvider *FakeProvider) WriteFile(bucket, path string, contents []byte) error {
	return fakeProvider.FakeWriteFile(bucket, path, contents)
}

// Zone is a fake zone
func (fakeProvider *FakeProvider) Zone(zone string) string {
	return fakeProvider.FakeZone(zone)
}

// DBType is a fake DBType
func (fakeProvider *FakeProvider) DBType(size string) string {
	return fakeProvider.FakeDBType(size)
}

// Choose is a fake Choose
func (fakeProvider *FakeProvider) Choose(c iaas.Choice) interface{} {
	return fakeProvider.FakeChoose(c)
}

// FakeTerraformInputVars implements terraform.TerraformInputVars for testing
type FakeTerraformInputVars struct {
	Key                    string
	FakeConfigureTerraform func(config string) (string, error)
	FakeBuild              func(data map[string]interface{}) error
}

// ConfigureTerraform delegates to FakeConfigureTerraform which is dynamically set by the tests
func (fakeTerraformInputVars *FakeTerraformInputVars) ConfigureTerraform(config string) (string, error) {
	return fakeTerraformInputVars.FakeConfigureTerraform(config)
}

// Build delegates to FakeBuild which is dynamically set by the tests
func (fakeTerraformInputVars *FakeTerraformInputVars) Build(data map[string]interface{}) error {
	return fakeTerraformInputVars.FakeBuild(data)
}

// FakeConfigClient implements config.IClient for testing
type FakeConfigClient struct {
	FakeLoad         func() (config.Config, error)
	FakeUpdate       func(config.Config) error
	FakeStoreAsset   func(filename string, contents []byte) error
	FakeLoadAsset    func(filename string) ([]byte, error)
	FakeDeleteAsset  func(filename string) error
	FakeDeleteAll    func(config config.Config) error
	FakeHasAsset     func(filename string) (bool, error)
	FakeNewConfig    func() config.Config
	FakeConfigExists func() (bool, error)
}

// Load delegates to FakeLoad which is dynamically set by the tests
func (client *FakeConfigClient) Load() (config.Config, error) {
	return client.FakeLoad()
}

// Update delegates to FakeUpdate which is dynamically set by the tests
func (client *FakeConfigClient) Update(config config.Config) error {
	return client.FakeUpdate(config)
}

// StoreAsset delegates to FakeStoreAsset which is dynamically set by the tests
func (client *FakeConfigClient) StoreAsset(filename string, contents []byte) error {
	return client.FakeStoreAsset(filename, contents)
}

// LoadAsset delegates to FakeLoadAsset which is dynamically set by the tests
func (client *FakeConfigClient) LoadAsset(filename string) ([]byte, error) {
	return client.FakeLoadAsset(filename)
}

// DeleteAsset delegates to FakeDeleteAsset which is dynamically set by the tests
func (client *FakeConfigClient) DeleteAsset(filename string) error {
	return client.FakeDeleteAsset(filename)
}

// DeleteAll delegates to FakeDeleteAll which is dynamically set by the tests
func (client *FakeConfigClient) DeleteAll(config config.Config) error {
	return client.FakeDeleteAll(config)
}

// HasAsset delegates to FakeHasAsset which is dynamically set by the tests
func (client *FakeConfigClient) HasAsset(filename string) (bool, error) {
	return client.FakeHasAsset(filename)
}

func (client *FakeConfigClient) NewConfig() config.Config {
	return client.NewConfig()
}

func (client *FakeConfigClient) ConfigExists() (bool, error) {
	return client.FakeConfigExists()
}

// FakeBoshClient implements bosh.IClient for testing
type FakeBoshClient struct {
	FakeDeploy    func([]byte, []byte, bool) ([]byte, []byte, error)
	FakeDelete    func([]byte) ([]byte, error)
	FakeCleanup   func() error
	FakeInstances func() ([]bosh.Instance, error)
	FakeCreateEnv func([]byte, []byte, string) ([]byte, []byte, error)
	FakeRecreate  func() error
	FakeLocks     func() ([]byte, error)
}

// Deploy delegates to FakeDeploy which is dynamically set by the tests
func (client *FakeBoshClient) Deploy(stateFileBytes, credsFileBytes []byte, detach bool) ([]byte, []byte, error) {
	return client.FakeDeploy(stateFileBytes, credsFileBytes, detach)
}

// Delete delegates to FakeDelete which is dynamically set by the tests
func (client *FakeBoshClient) Delete(stateFileBytes []byte) ([]byte, error) {
	return client.FakeDelete(stateFileBytes)
}

// Cleanup delegates to FakeCleanup which is dynamically set by the tests
func (client *FakeBoshClient) Cleanup() error {
	return client.FakeCleanup()
}

// Instances delegates to FakeInstances which is dynamically set by the tests
func (client *FakeBoshClient) Instances() ([]bosh.Instance, error) {
	return client.FakeInstances()
}

// CreateEnv delegates to FakeCreateEnv which is dynamically set by the tests
func (client *FakeBoshClient) CreateEnv(state, creds []byte, customOps string) ([]byte, []byte, error) {
	return client.FakeCreateEnv(state, creds, customOps)
}

// Recreate delegates to FakeRecreate which is dynamically set by the tests
func (client *FakeBoshClient) Recreate() error {
	return client.FakeRecreate()
}

// Locks delegates to FakeLocks which is dynamically set by the tests
func (client *FakeBoshClient) Locks() ([]byte, error) {
	return client.FakeLocks()
}

// NewFakeAcmeClient returns a new FakeAcmeClient
func NewFakeAcmeClient(u *certs.User) (*lego.Client, error) {
	return &lego.Client{}, nil
}
