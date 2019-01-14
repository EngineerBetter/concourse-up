package testsupport

import (
	"bytes"
	"crypto"
	"errors"

	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/terraform"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/xenolf/lego/acme"
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
	FakeIAAS                          func() string
	FakeLoadFile                      func(bucket, path string) ([]byte, error)
	FakeRegion                        func() string
	FakeWorkerType                    func(string)
	FakeWriteFile                     func(bucket, path string, contents []byte) error
	FakeZone                          func(string) string
	FakeDBType                        func(string) string
}

// DBType is a fake DBType
func (fakeProvider *FakeProvider) DBType(size string) string {
	return fakeProvider.FakeDBType(size)
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
func (fakeProvider *FakeProvider) IAAS() string {
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

// FakeTerraformInputVars implements terraform.TerraformInputVars for testing
type FakeTerraformInputVars struct {
	FakeConfigureTerraform func(config string) (string, error)
	FakeBuild              func(data map[string]interface{}) error
}

// FakeIAASMetadata implements terraform.IAASMetadata for testing
type FakeIAASMetadata struct {
	FakeAssertValid func() error
	FakeInit        func(*bytes.Buffer) error
	FakeGet         func(string) (string, error)
}

// AssertValid delegates to FakeAssertValid which is dynamically set by the tests
func (fakeIAASMetadata *FakeIAASMetadata) AssertValid() error {
	return fakeIAASMetadata.FakeAssertValid()
}

// Init delegates to FakeInit which is dynamically set by the tests
func (fakeIAASMetadata *FakeIAASMetadata) Init(data *bytes.Buffer) error {
	return fakeIAASMetadata.FakeInit(data)
}

// Get delegates to FakeGet which is dynamically set by the tests
func (fakeIAASMetadata *FakeIAASMetadata) Get(key string) (string, error) {
	return fakeIAASMetadata.FakeGet(key)
}

// ConfigureTerraform delegates to FakeConfigureTerraform which is dynamically set by the tests
func (fakeTerraformInputVars *FakeTerraformInputVars) ConfigureTerraform(config string) (string, error) {
	return fakeTerraformInputVars.FakeConfigureTerraform(config)
}

// Build delegates to FakeBuild which is dynamically set by the tests
func (fakeTerraformInputVars *FakeTerraformInputVars) Build(data map[string]interface{}) error {
	return fakeTerraformInputVars.FakeBuild(data)
}

// FakeCLI implements terraform.CLI for testing
type FakeCLI struct {
	FakeIAAS        func(name string) (terraform.InputVars, terraform.IAASMetadata, error)
	FakeApply       func(conf terraform.InputVars, dryrun bool) error
	FakeDestroy     func(conf terraform.InputVars) error
	FakeBuildOutput func(conf terraform.InputVars, metadata terraform.IAASMetadata) error
}

// IAAS delegates to FakeIAAS which is dynamically set by the tests
func (client *FakeCLI) IAAS(name string) (terraform.InputVars, terraform.IAASMetadata, error) {
	return client.FakeIAAS(name)
}

// Apply delegates to FakeApply which is dynamically set by the tests
func (client *FakeCLI) Apply(conf terraform.InputVars, dryrun bool) error {
	return client.FakeApply(conf, dryrun)
}

// Destroy delegates to FakeDestroy which is dynamically set by the tests
func (client *FakeCLI) Destroy(conf terraform.InputVars) error {
	return client.FakeDestroy(conf)
}

// BuildOutput delegates to FakeBuildOutput which is dynamically set by the tests
func (client *FakeCLI) BuildOutput(conf terraform.InputVars, metadata terraform.IAASMetadata) error {
	return client.FakeBuildOutput(conf, metadata)
}

// FakeAWSClient implements iaas.IClient for testing
type FakeAWSClient struct {
	FakeCheckForWhitelistedIP         func(ip, securityGroup string) (bool, error)
	FakeDeleteVMsInVPC                func(vpcID string) ([]string, error)
	FakeDeleteVolumes                 func(volumesToDelete []string, deleteVolume func(ec2Client iaas.IEC2, volumeID *string) error) error
	FakeDeleteFile                    func(bucket, path string) error
	FakeDeleteVersionedBucket         func(name string) error
	FakeCreateBucket                  func(name string) error
	FakeBucketExists                  func(name string) (bool, error)
	FakeEnsureFileExists              func(bucket, path string, defaultContents []byte) ([]byte, bool, error)
	FakeFindLongestMatchingHostedZone func(subdomain string) (string, string, error)
	FakeHasFile                       func(bucket, path string) (bool, error)
	FakeLoadFile                      func(bucket, path string) ([]byte, error)
	FakeNewEC2Client                  func() (iaas.IEC2, error)
	FakeWriteFile                     func(bucket, path string, contents []byte) error
	FakeRegion                        func() string
	FakeAttr                          func(string) (string, error)
	FakeZone                          func(string) string
	FakeDeleteVMsInDeployment         func(zone, project, deployment string) error
	FakeWorkerType                    func(string)
	FakeCreateDatabases               func(name, username, password string) error
	FakeDBType                        func(size string) string
}

// DBType is here to implement iaas.DBType
func (client *FakeAWSClient) DBType(size string) string {
	return client.FakeDBType(size)
}

// WorkerType is here to implement iaas.IClient
func (client *FakeAWSClient) WorkerType(w string) {
}

// IAAS is here to implement iaas.IClient
func (client *FakeAWSClient) IAAS() string {
	return "AWS"
}

// DeleteVMsInDeployment delegates to FakeDeleteVMsInDeployment which is dynamically set by the tests
func (client *FakeAWSClient) DeleteVMsInDeployment(zone, project, deployment string) error {
	return client.FakeDeleteVMsInDeployment(zone, project, deployment)
}

// Attr is here to implement iaas.IClient
func (client *FakeAWSClient) Attr(a string) (string, error) {
	return client.FakeAttr(a)
}

// Zone is here to implement iaas.IClient
func (client *FakeAWSClient) Zone(z string) string {
	return client.FakeZone(z)
}

// Region delegates to FakeRegion which is dynamically set by the tests
func (client *FakeAWSClient) Region() string {
	return client.FakeRegion()
}

// CheckForWhitelistedIP delegates to FakeCheckForWhitelistedIP which is dynamically set by the tests
func (client *FakeAWSClient) CheckForWhitelistedIP(ip, securityGroup string) (bool, error) {
	return client.FakeCheckForWhitelistedIP(ip, securityGroup)
}

// DeleteVMsInVPC delegates to FakeDeleteVMsInVPC which is dynamically set by the tests
func (client *FakeAWSClient) DeleteVMsInVPC(vpcID string) ([]string, error) {
	return client.FakeDeleteVMsInVPC(vpcID)
}

// DeleteVolumes delegates to FakeDeleteVolumes which is dynamically set by the tests
func (client *FakeAWSClient) DeleteVolumes(volumesToDelete []string, deleteVolume func(ec2Client iaas.IEC2, volumeID *string) error) error {
	return client.FakeDeleteVolumes(volumesToDelete, deleteVolume)
}

// DeleteFile delegates to FakeDeleteFile which is dynamically set by the tests
func (client *FakeAWSClient) DeleteFile(bucket, path string) error {
	return client.FakeDeleteFile(bucket, path)
}

// DeleteVersionedBucket delegates to FakeDeleteVersionedBucket which is dynamically set by the tests
func (client *FakeAWSClient) DeleteVersionedBucket(name string) error {
	return client.FakeDeleteVersionedBucket(name)
}

// CreateBucket delegates to FakeCreateBucket which is dynamically set by the tests
func (client *FakeAWSClient) CreateBucket(name string) error {
	return client.FakeCreateBucket(name)
}

// BucketExists delegates to FakeBucketExists which is dynamically set by the tests
func (client *FakeAWSClient) BucketExists(name string) (bool, error) {
	return client.FakeBucketExists(name)
}

// EnsureFileExists delegates to FakeEnsureFileExists which is dynamically set by the tests
func (client *FakeAWSClient) EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
	return client.FakeEnsureFileExists(bucket, path, defaultContents)
}

// FindLongestMatchingHostedZone delegates to FakeFindLongestMatchingHostedZone which is dynamically set by the tests
func (client *FakeAWSClient) FindLongestMatchingHostedZone(subdomain string) (string, string, error) {
	return client.FakeFindLongestMatchingHostedZone(subdomain)
}

// HasFile delegates to FakeHasFile which is dynamically set by the tests
func (client *FakeAWSClient) HasFile(bucket, path string) (bool, error) {
	return client.FakeHasFile(bucket, path)
}

// LoadFile delegates to FakeLoadFile which is dynamically set by the tests
func (client *FakeAWSClient) LoadFile(bucket, path string) ([]byte, error) {
	return client.FakeLoadFile(bucket, path)
}

// NewEC2Client delegates to FakeNewEC2Client which is dynamically set by the tests
func (client *FakeAWSClient) NewEC2Client() (iaas.IEC2, error) {
	return client.FakeNewEC2Client()
}

// WriteFile delegates to FakeWriteFile which is dynamically set by the tests
func (client *FakeAWSClient) WriteFile(bucket, path string, contents []byte) error {
	return client.FakeWriteFile(bucket, path, contents)
}

// CreateDatabases delegates to FakeCreateDatabases which is dynamically set by the tests
func (client *FakeAWSClient) CreateDatabases(name, username, password string) error {
	return client.FakeCreateDatabases(name, username, password)
}

// FakeFlyClient implements fly.IClient for testing
type FakeFlyClient struct {
	FakeSetDefaultPipeline func(config config.Config, allowFlyVersionDiscrepancy bool) error
	FakeCleanup            func() error
	FakeCanConnect         func() (bool, error)
}

// SetDefaultPipeline delegates to FakeSetDefaultPipeline which is dynamically set by the tests
func (client *FakeFlyClient) SetDefaultPipeline(config config.Config, allowFlyVersionDiscrepancy bool) error {
	return client.FakeSetDefaultPipeline(config, allowFlyVersionDiscrepancy)
}

// Cleanup delegates to FakeCleanup which is dynamically set by the tests
func (client *FakeFlyClient) Cleanup() error {
	return client.FakeCleanup()
}

// CanConnect delegates to FakeCanConnect which is dynamically set by the tests
func (client *FakeFlyClient) CanConnect() (bool, error) {
	return client.FakeCanConnect()
}

// FakeConfigClient implements config.IClient for testing
type FakeConfigClient struct {
	FakeLoad         func() (config.Config, error)
	FakeUpdate       func(config.Config) error
	FakeLoadOrCreate func(deployArgs *deploy.Args) (config.Config, bool, bool, error)
	FakeStoreAsset   func(filename string, contents []byte) error
	FakeLoadAsset    func(filename string) ([]byte, error)
	FakeDeleteAsset  func(filename string) error
	FakeDeleteAll    func(config config.Config) error
	FakeHasAsset     func(filename string) (bool, error)
}

// Load delegates to FakeLoad which is dynamically set by the tests
func (client *FakeConfigClient) Load() (config.Config, error) {
	return client.FakeLoad()
}

// Update delegates to FakeUpdate which is dynamically set by the tests
func (client *FakeConfigClient) Update(config config.Config) error {
	return client.FakeUpdate(config)
}

// LoadOrCreate delegates to FakeLoadOrCreate which is dynamically set by the tests
func (client *FakeConfigClient) LoadOrCreate(deployArgs *deploy.Args) (config.Config, bool, bool, error) {
	return client.FakeLoadOrCreate(deployArgs)
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

// FakeBoshClient implements bosh.IClient for testing
type FakeBoshClient struct {
	FakeDeploy    func([]byte, []byte, bool) ([]byte, []byte, error)
	FakeDelete    func([]byte) ([]byte, error)
	FakeCleanup   func() error
	FakeInstances func() ([]bosh.Instance, error)
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

// NewFakeAcmeClient returns a new FakeAcmeClient
func NewFakeAcmeClient(u *certs.User) (certs.AcmeClient, error) {
	return &FakeAcmeClient{}, nil
}

// FakeAcmeClient implements certs.AcmeClient for testing
type FakeAcmeClient struct {
}

// SetChallengeProvider returns nil
func (c *FakeAcmeClient) SetChallengeProvider(challenge acme.Challenge, p acme.ChallengeProvider) error {
	return nil
}

// ExcludeChallenges does nothing
func (c *FakeAcmeClient) ExcludeChallenges(challenges []acme.Challenge) {
}

// Register returns nil
func (c *FakeAcmeClient) Register() (*acme.RegistrationResource, error) {
	return nil, nil
}

// AgreeToTOS returns nil
func (c *FakeAcmeClient) AgreeToTOS() error {
	return nil
}

// ObtainCertificate returns a fake certificate if domain is valid
func (c *FakeAcmeClient) ObtainCertificate(domains []string, bundle bool, privKey crypto.PrivateKey, mustStaple bool) (acme.CertificateResource, map[string]error) {
	if domains[0] == "google.com" {
		errs := make(map[string]error)
		errs["error"] = errors.New("this is an error")
		return acme.CertificateResource{}, errs
	}
	return acme.CertificateResource{
		PrivateKey:        []byte("BEGIN RSA PRIVATE KEY"),
		Certificate:       []byte("BEGIN CERTIFICATE"),
		IssuerCertificate: []byte("BEGIN CERTIFICATE"),
	}, nil
}
