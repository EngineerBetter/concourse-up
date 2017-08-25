package concourse_test

import (
	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConcourse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Concourse Suite")
}

type FakeAWSClient struct {
	FakeDeleteVMsInVPC                func(vpcID string) error
	FakeDeleteFile                    func(bucket, path string) error
	FakeDeleteVersionedBucket         func(name string) error
	FakeEnsureBucketExists            func(name string) error
	FakeEnsureFileExists              func(bucket, path string, defaultContents []byte) ([]byte, bool, error)
	FakeFindLongestMatchingHostedZone func(subdomain string) (string, string, error)
	FakeHasFile                       func(bucket, path string) (bool, error)
	FakeLoadFile                      func(bucket, path string) ([]byte, error)
	FakeWriteFile                     func(bucket, path string, contents []byte) error
	FakeRegion                        func() string
}

func (client *FakeAWSClient) IAAS() string {
	return "AWS"
}

func (client *FakeAWSClient) Region() string {
	return client.FakeRegion()
}
func (client *FakeAWSClient) DeleteVMsInVPC(vpcID string) error {
	return client.FakeDeleteVMsInVPC(vpcID)
}
func (client *FakeAWSClient) DeleteFile(bucket, path string) error {
	return client.FakeDeleteFile(bucket, path)
}
func (client *FakeAWSClient) DeleteVersionedBucket(name string) error {
	return client.FakeDeleteVersionedBucket(name)
}
func (client *FakeAWSClient) EnsureBucketExists(name string) error {
	return client.FakeEnsureBucketExists(name)
}
func (client *FakeAWSClient) EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
	return client.FakeEnsureFileExists(bucket, path, defaultContents)
}
func (client *FakeAWSClient) FindLongestMatchingHostedZone(subdomain string) (string, string, error) {
	return client.FakeFindLongestMatchingHostedZone(subdomain)
}
func (client *FakeAWSClient) HasFile(bucket, path string) (bool, error) {
	return client.FakeHasFile(bucket, path)
}
func (client *FakeAWSClient) LoadFile(bucket, path string) ([]byte, error) {
	return client.FakeLoadFile(bucket, path)
}
func (client *FakeAWSClient) WriteFile(bucket, path string, contents []byte) error {
	return client.FakeWriteFile(bucket, path, contents)
}

type FakeFlyClient struct {
	FakeSetDefaultPipeline func(deployAgs *config.DeployArgs, config *config.Config) error
	FakeCleanup            func() error
	FakeCanConnect         func() (bool, error)
}

func (client *FakeFlyClient) SetDefaultPipeline(deployArgs *config.DeployArgs, config *config.Config) error {
	return client.FakeSetDefaultPipeline(deployArgs, config)
}

func (client *FakeFlyClient) Cleanup() error {
	return client.FakeCleanup()
}

func (client *FakeFlyClient) CanConnect() (bool, error) {
	return client.FakeCanConnect()
}

type FakeConfigClient struct {
	FakeLoad         func() (*config.Config, error)
	FakeUpdate       func(*config.Config) error
	FakeLoadOrCreate func(deployArgs *config.DeployArgs) (*config.Config, bool, error)
	FakeStoreAsset   func(filename string, contents []byte) error
	FakeLoadAsset    func(filename string) ([]byte, error)
	FakeDeleteAsset  func(filename string) error
	FakeDeleteAll    func(config *config.Config) error
	FakeHasAsset     func(filename string) (bool, error)
}

func (client *FakeConfigClient) Load() (*config.Config, error) {
	return client.FakeLoad()
}

func (client *FakeConfigClient) Update(config *config.Config) error {
	return client.FakeUpdate(config)
}

func (client *FakeConfigClient) LoadOrCreate(deployArgs *config.DeployArgs) (*config.Config, bool, error) {
	return client.FakeLoadOrCreate(deployArgs)
}

func (client *FakeConfigClient) StoreAsset(filename string, contents []byte) error {
	return client.FakeStoreAsset(filename, contents)
}

func (client *FakeConfigClient) LoadAsset(filename string) ([]byte, error) {
	return client.FakeLoadAsset(filename)
}

func (client *FakeConfigClient) DeleteAsset(filename string) error {
	return client.FakeDeleteAsset(filename)
}

func (client *FakeConfigClient) DeleteAll(config *config.Config) error {
	return client.FakeDeleteAll(config)
}

func (client *FakeConfigClient) HasAsset(filename string) (bool, error) {
	return client.FakeHasAsset(filename)
}

type FakeTerraformClient struct {
	FakeOutput  func() (*terraform.Metadata, error)
	FakeApply   func(dryrun bool) error
	FakeDestroy func() error
	FakeCleanup func() error
}

func (client *FakeTerraformClient) Output() (*terraform.Metadata, error) {
	return client.FakeOutput()
}

func (client *FakeTerraformClient) Apply(dryrun bool) error {
	return client.FakeApply(dryrun)
}

func (client *FakeTerraformClient) Destroy() error {
	return client.FakeDestroy()
}

func (client *FakeTerraformClient) Cleanup() error {
	return client.FakeCleanup()
}

type FakeBoshClient struct {
	FakeDeploy    func([]byte, bool) ([]byte, error)
	FakeDelete    func([]byte) ([]byte, error)
	FakeCleanup   func() error
	FakeInstances func() ([]bosh.Instance, error)
}

func (client *FakeBoshClient) Deploy(stateFileBytes []byte, detach bool) ([]byte, error) {
	return client.FakeDeploy(stateFileBytes, detach)
}

func (client *FakeBoshClient) Delete(stateFileBytes []byte) ([]byte, error) {
	return client.FakeDelete(stateFileBytes)
}

func (client *FakeBoshClient) Cleanup() error {
	return client.FakeCleanup()
}

func (client *FakeBoshClient) Instances() ([]bosh.Instance, error) {
	return client.FakeInstances()
}
