package config_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	. "github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/testsupport"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Client", func() {
	var iaasClient *testsupport.FakeAWSClient
	var client *Client
	var deployArgs *DeployArgs

	BeforeEach(func() {
		iaasClient = &testsupport.FakeAWSClient{
			FakeRegion: func() string {
				return "eu-west-1"
			},
			FakeCreateBucket: func(name string) error {
				return nil
			},
			FakeBucketExists: func(name string) (bool, error) {
				return false, nil
			},
			FakeEnsureFileExists: func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
				return defaultContents, true, nil
			},
		}
		client = New(iaasClient, "test", "")

		deployArgs = &DeployArgs{
			IAAS:        "AWS",
			AWSRegion:   "eu-west-1",
			WorkerCount: 1,
			WorkerSize:  "xlarge",
			DBSize:      "medium",
			DBSizeIsSet: false,
			AllowIPs:    "0.0.0.0",
			Spot:        true,
		}
	})

	DescribeTable("parseCDIRBlocks",
		func(in, out string) {
			b, err := ParseCIDRBlocks(in)
			Expect(err).NotTo(HaveOccurred())
			Expect(b.String()).To(Equal(out))
		},
		Entry("Single IP", "8.8.8.8", `"8.8.8.8/32"`),
		Entry("Single CIDR Block", "1.2.3.0/28", `"1.2.3.0/28"`),
		Entry("IP and CIDR Block", "8.8.8.8,1.2.3.0/28", `"8.8.8.8/32", "1.2.3.0/28"`),
	)

	Describe("LoadOrCreate", func() {
		Context("When the there is no existing config", func() {
			var conf Config
			var createdANewFile bool

			BeforeEach(func() {
				var err error
				conf, createdANewFile, err = client.LoadOrCreate(deployArgs)
				Expect(err).To(Succeed())
			})

			It("creates a new file", func() {
				Expect(createdANewFile).To(BeTrue())
			})

			Describe("the default config file", func() {
				It("Sets the default value for the AvailabilityZone", func() {
					Expect(conf.AvailabilityZone).To(Equal("eu-west-1a"))
				})

				It("Sets the default value for the ConcourseDBName", func() {
					Expect(conf.ConcourseDBName).To(Equal("concourse_atc"))
				})

				It("Sets the default value for the ConcourseWorkerCount", func() {
					Expect(conf.ConcourseWorkerCount).To(Equal(1))
				})

				It("Sets the default value for the ConcourseWorkerSize", func() {
					Expect(conf.ConcourseWorkerSize).To(Equal("xlarge"))
				})

				It("Sets the default value for the ConfigBucket", func() {
					Expect(conf.ConfigBucket).To(Equal("concourse-up-test-eu-west-1-config"))
				})

				It("Sets the default value for the Deployment", func() {
					Expect(conf.Deployment).To(Equal("concourse-up-test"))
				})

				It("Generates a secure random string for DirectorHMUserPassword", func() {
					Expect(conf.DirectorHMUserPassword).To(beARandomPassword())
				})

				It("Generates a secure random string for DirectorMbusPassword", func() {
					Expect(conf.DirectorMbusPassword).To(beARandomPassword())
				})

				It("Generates a secure random string for DirectorNATSPassword", func() {
					Expect(conf.DirectorNATSPassword).To(beARandomPassword())
				})

				It("Generates a secure random string for DirectorPassword", func() {
					Expect(conf.DirectorPassword).To(beARandomPassword())
				})

				It("Generates a secure random string for DirectorRegistryPassword", func() {
					Expect(conf.DirectorRegistryPassword).To(beARandomPassword())
				})

				It("Sets the default value for the DirectorUsername", func() {
					Expect(conf.DirectorUsername).To(Equal("admin"))
				})

				It("Generates a secure random string for EncryptionKey", func() {
					Expect(conf.EncryptionKey).To(MatchRegexp("^[a-z0-9]{32}$"))
				})

				It("Sets the GrafanaPassword to the ConcoursePassword", func() {
					Expect(conf.GrafanaPassword).To(Equal(conf.ConcoursePassword))
				})

				It("Sets the default value for the MultiAZRDS", func() {
					Expect(conf.MultiAZRDS).To(Equal(false))
				})

				It("Generates a random RSA private key for PrivateKey", func() {
					Expect(conf.PrivateKey).To(HavePrefix("-----BEGIN RSA PRIVATE KEY-----"))
				})

				It("Sets the default value for the Project", func() {
					Expect(conf.Project).To(Equal("test"))
				})

				It("Generates a random RSA public key for PublicKey", func() {
					Expect(conf.PublicKey).To(HavePrefix("ssh-rsa"))
				})

				It("Sets the default value for the RDSDefaultDatabaseName", func() {
					Expect(conf.RDSDefaultDatabaseName).To(Equal("bosh"))
				})

				It("Sets the default value for the RDSInstanceClass", func() {
					Expect(conf.RDSInstanceClass).To(Equal("db.t2.medium"))
				})

				It("Generates a secure random string for RDSPassword", func() {
					Expect(conf.RDSPassword).To(beARandomPassword())
				})

				It("Generates a secure random string for the RDSUsername", func() {
					Expect(conf.RDSUsername).To(MatchRegexp("^admin[a-z0-9]{20}$"))
				})

				It("Sets the default value for the Region", func() {
					Expect(conf.Region).To(Equal("eu-west-1"))
				})

				It("Sets the default value for the TFStatePath", func() {
					Expect(conf.TFStatePath).To(Equal("terraform.tfstate"))
				})

				It("Sets the Spot field", func() {
					Expect(conf.Spot).To(Equal(true))
				})
			})
			Describe("github auth flags are present", func() {
				It("sets the github auth config fields", func() {
					deployArgs.ModifyGithub("id", "secret", true)
					conf, createdANewFile, err := client.LoadOrCreate(deployArgs)
					Expect(err).To(Succeed())
					Expect(createdANewFile).To(BeTrue())
					Expect(conf.GithubClientID).To(Equal("id"))
					Expect(conf.GithubClientSecret).To(Equal("secret"))
					Expect(conf.GithubAuthIsSet).To(BeTrue())
				})
			})
		})
	})
})

func beARandomPassword() types.GomegaMatcher {
	return MatchRegexp("^[a-z0-9]{20}$")
}

func TestNew(t *testing.T) {
	var iaasClient *testsupport.FakeAWSClient
	iaasClient = &testsupport.FakeAWSClient{
		FakeRegion: func() string {
			return "eu-west-1"
		},
		FakeCreateBucket: func(name string) error {
			return nil
		},
		FakeEnsureFileExists: func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
			return defaultContents, true, nil
		},
	}

	type args struct {
		iaas      iaas.IClient
		project   string
		namespace string
	}
	tests := []struct {
		name             string
		args             args
		want             *Client
		FakeBucketExists func(name string) (bool, error)
	}{
		{
			name: "default",
			args: args{
				iaas:      iaasClient,
				project:   "aProject",
				namespace: "",
			},
			want: &Client{
				Iaas:         iaasClient,
				Project:      "aProject",
				Namespace:    "eu-west-1",
				BucketName:   "concourse-up-aProject-eu-west-1-config",
				BucketExists: false,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				return false, nil
			},
		},
		{
			name: "with Namespace",
			args: args{
				iaas:      iaasClient,
				project:   "aProject",
				namespace: "someNamespace",
			},
			want: &Client{
				Iaas:         iaasClient,
				Project:      "aProject",
				Namespace:    "someNamespace",
				BucketName:   "concourse-up-aProject-someNamespace-config",
				BucketExists: false,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				return false, nil
			},
		},
		{
			name: "with Namespace and region based bucket",
			args: args{
				iaas:      iaasClient,
				project:   "aProject",
				namespace: "someNamespace",
			},
			want: &Client{
				Iaas:         iaasClient,
				Project:      "aProject",
				Namespace:    "someNamespace",
				BucketName:   "concourse-up-aProject-eu-west-1-config",
				BucketExists: true,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				if name == "concourse-up-aProject-eu-west-1-config" {
					return true, nil
				}
				return false, nil
			},
		},
		{
			name: "with Namespace and namespace based bucket",
			args: args{
				iaas:      iaasClient,
				project:   "aProject",
				namespace: "someNamespace",
			},
			want: &Client{
				Iaas:         iaasClient,
				Project:      "aProject",
				Namespace:    "someNamespace",
				BucketName:   "concourse-up-aProject-someNamespace-config",
				BucketExists: true,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				if name == "concourse-up-aProject-someNamespace-config" {
					return true, nil
				}
				return false, nil
			},
		},
		{
			name: "with Namespace and both buckets",
			args: args{
				iaas:      iaasClient,
				project:   "aProject",
				namespace: "someNamespace",
			},
			want: &Client{
				Iaas:         iaasClient,
				Project:      "aProject",
				Namespace:    "someNamespace",
				BucketName:   "",
				BucketExists: true,
				BucketError:  fmt.Errorf("found both region %q and namespaced %q buckets for %q deployment", "concourse-up-aProject-eu-west-1-config", "concourse-up-aProject-someNamespace-config", "aProject"),
			},
			FakeBucketExists: func(name string) (bool, error) {
				return true, nil
			},
		},
		{
			name: "with Namespace and bucket existing and namespace == region",
			args: args{
				iaas:      iaasClient,
				project:   "aProject",
				namespace: "eu-west-1",
			},
			want: &Client{
				Iaas:         iaasClient,
				Project:      "aProject",
				Namespace:    "eu-west-1",
				BucketName:   "concourse-up-aProject-eu-west-1-config",
				BucketExists: true,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				return true, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iaasClient.FakeBucketExists = tt.FakeBucketExists
			if got := New(tt.args.iaas, tt.args.project, tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v,\n want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Load(t *testing.T) {
	var iaasClient *testsupport.FakeAWSClient
	iaasClient = &testsupport.FakeAWSClient{
		FakeRegion: func() string {
			return "eu-west-1"
		},
		FakeCreateBucket: func(name string) error {
			return nil
		},
		FakeEnsureFileExists: func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
			return defaultContents, true, nil
		},
		FakeLoadFile: func(bucket, path string) ([]byte, error) {
			bytes, _ := json.Marshal(Config{})
			return bytes, nil
		},
	}
	tests := []struct {
		name    string
		prepare func() *Client
		want    Config
		wantErr bool
	}{
		{
			name: "BucketError is raised",
			prepare: func() *Client {
				return &Client{
					Iaas:         iaasClient,
					Project:      "",
					Namespace:    "",
					BucketName:   "",
					BucketExists: false,
					BucketError:  fmt.Errorf("some error"),
				}
			},
			want:    Config{},
			wantErr: true,
		},
		{
			name: "default",
			prepare: func() *Client {
				return &Client{
					Iaas:         iaasClient,
					Project:      "",
					Namespace:    "",
					BucketName:   "",
					BucketExists: false,
					BucketError:  nil,
				}
			},
			want:    Config{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.prepare()
			got, err := client.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Load() = %v, want %v", got, tt.want)
			}
		})
	}
}
