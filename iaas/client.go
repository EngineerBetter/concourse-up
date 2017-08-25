package iaas

// IClient represents actions taken against AWS
type IClient interface {
	DeleteFile(bucket, path string) error
	DeleteVersionedBucket(name string) error
	DeleteVMsInVPC(vpcID string) error
	EnsureBucketExists(name string) error
	EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error)
	FindLongestMatchingHostedZone(subdomain string) (string, string, error)
	HasFile(bucket, path string) (bool, error)
	LoadFile(bucket, path string) ([]byte, error)
	WriteFile(bucket, path string, contents []byte) error
	Region() string
	IAAS() string
}
