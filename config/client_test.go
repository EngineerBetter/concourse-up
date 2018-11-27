package config_test

import (
	"encoding/json"
	"errors"
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
				conf, createdANewFile, _, err = client.LoadOrCreate(deployArgs)
				Expect(err).To(Succeed())
			})

			It("creates a new file", func() {
				Expect(createdANewFile).To(BeTrue())
			})

			Describe("the default config file", func() {
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
					conf, createdANewFile, _, err := client.LoadOrCreate(deployArgs)
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
		iaas      iaas.Provider
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
				Config:       &Config{},
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
				Config:       &Config{},
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
				Config:       &Config{},
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
				Config:       &Config{},
			},
			FakeBucketExists: func(name string) (bool, error) {
				if name == "concourse-up-aProject-someNamespace-config" {
					return true, nil
				}
				return false, nil
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
				Config:       &Config{},
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
					Config:       &Config{},
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
					Config:       &Config{},
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

func TestClient_LoadOrCreate(t *testing.T) {
	type fields struct {
		Iaas         *testsupport.FakeAWSClient
		Project      string
		Namespace    string
		BucketName   string
		BucketExists bool
		BucketError  error
		Config       *Config
	}
	type args struct {
		deployArgs *DeployArgs
	}
	type returnVals struct {
		config           Config
		newConfigCreated bool
		isDomainUpdated  bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		validate func(Config, bool, bool) (bool, string)
	}{
		{
			name: "failure- bucket error",
			fields: fields{
				BucketError: errors.New("a bucket error"),
			},
			validate: func(c Config, newConfigCreated, isDomainUpdated bool) (bool, string) {
				expected := returnVals{Config{}, false, false}
				got := returnVals{c, newConfigCreated, isDomainUpdated}
				isValid := reflect.DeepEqual(expected, got)
				return isValid, fmt.Sprintf("expected: %v\n received: %v\n", expected, got)
			},
			wantErr: true,
		},
		{
			name: "first time deployment without flags- use deploy args to set config to defaults",
			fields: fields{
				Iaas: &testsupport.FakeAWSClient{
					FakeEnsureFileExists: func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
						return defaultContents, true, nil
					},
					FakeBucketExists: func(name string) (bool, error) {
						return true, nil
					},
					FakeRegion: func() string {
						return "eu-west-1"
					},
				},
			},
			args: args{
				&DeployArgs{
					AllowIPs:    "0.0.0.0",
					WorkerCount: 1,
					WorkerSize:  "xlarge",
					WebSize:     "small",
					DBSize:      "small",
					Domain:      "FakeDomain",
				},
			},
			wantErr: false,
			validate: func(c Config, newConfigCreated, isDomainUpdated bool) (bool, string) {
				isValidConfig :=
					c.ConcourseWorkerCount == 1 &&
						c.ConcourseWorkerSize == "xlarge" &&
						c.ConcourseWebSize == "small" &&
						c.RDSInstanceClass == "db.t2.small" &&
						c.Spot == false &&
						c.Domain == "FakeDomain"
				isValid := isValidConfig && newConfigCreated
				message := fmt.Sprintf("Config Key | Expected | \tReceived |\nConcourseWorkerCount| %v |\t%v|\nConcourseWorkerSize | %v |\t%v |\nConcourseWebSize | %v |\t %v |\nRDSInstanceClass |%v |\t| %v |\n|Spot | %v |\t| %v |", 1, c.ConcourseWorkerCount, "xlarge", c.ConcourseWorkerSize, "small", c.ConcourseWebSize, "db.t2.small", c.RDSInstanceClass, false, c.Spot)
				return isValid, message
			},
		},
		{
			name: "a subsequent deployment retains previous config values",
			fields: fields{
				Iaas: &testsupport.FakeAWSClient{
					FakeEnsureFileExists: func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
						fakeConfig := Config{
							Region:               "eu-west-3",
							ConcourseWorkerCount: 2,
							ConcourseWebSize:     "large",
							ConcourseWorkerSize:  "medium",
							RDSInstanceClass:     "db.m4.4xlarge",
							Domain:               "FakeDomain",
						}
						jsonConfig, _ := json.Marshal(fakeConfig)
						return jsonConfig, false, nil
					},
					FakeBucketExists: func(name string) (bool, error) {
						return true, nil
					},
					FakeRegion: func() string {
						return "eu-west-3"
					},
				},
			},
			args: args{
				&DeployArgs{
					AllowIPs: "0.0.0.0",
				},
			},
			wantErr: false,
			validate: func(c Config, newConfigCreated, isDomainUpdated bool) (bool, string) {
				isValidConfig :=
					c.ConcourseWorkerCount == 2 &&
						c.ConcourseWorkerSize == "medium" &&
						c.ConcourseWebSize == "large" &&
						c.Region == "eu-west-3" &&
						c.RDSInstanceClass == "db.m4.4xlarge" &&
						c.Domain == "FakeDomain"
				isValid := isValidConfig && !newConfigCreated && !isDomainUpdated
				message := fmt.Sprintf("Config Key | Expected | \tReceived |\nConcourseWorkerCount| %v |\t%v|\nConcourseWorkerSize | %v |\t%v |\nConcourseWebSize | %v |\t%v \nRegion | %v | %v |\nRDSInstanceClass | %v | %v|\t", 2, c.ConcourseWorkerCount, "medium", c.ConcourseWorkerSize, "large", c.ConcourseWebSize, "eu-west-3", c.Region, "db.m4.4xlarge", c.RDSInstanceClass)
				return isValid, message
			},
		},
		{
			name: "default config can be overriden in a subsequent deployment by passing different values via flags",
			fields: fields{
				Iaas: &testsupport.FakeAWSClient{
					FakeEnsureFileExists: func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
						fakeConfig := Config{
							Region:               "eu-west-3",
							ConcourseWorkerCount: 2,
							ConcourseWebSize:     "large",
							ConcourseWorkerSize:  "medium",
							RDSInstanceClass:     "db.m4.4xlarge",
							Domain:               "FakeDomainOne",
						}
						jsonConfig, _ := json.Marshal(fakeConfig)
						return jsonConfig, false, nil
					},
					FakeBucketExists: func(name string) (bool, error) {
						return true, nil
					},
					FakeRegion: func() string {
						return "eu-west-3"
					},
				},
			},
			args: args{
				&DeployArgs{
					AllowIPs:         "0.0.0.0",
					WorkerCountIsSet: true,
					WorkerCount:      1,
					DBSizeIsSet:      true,
					DBSize:           "small",
					Domain:           "FakeDomainTwo",
					DomainIsSet:      true,
				},
			},
			wantErr: false,
			validate: func(c Config, newConfigCreated, isDomainUpdated bool) (bool, string) {
				isValidConfig :=
					c.ConcourseWorkerCount == 1 &&
						c.ConcourseWorkerSize == "medium" &&
						c.ConcourseWebSize == "large" &&
						c.RDSInstanceClass == "db.t2.small" &&
						c.Region == "eu-west-3" &&
						c.Domain == "FakeDomainTwo"
				isValid := isValidConfig && !newConfigCreated && isDomainUpdated
				message := fmt.Sprintf("Config Key | Expected | \tReceived |\nConcourseWorkerCount| %v |\t%v|\nConcourseWorkerSize | %v |\t%v |\nConcourseWebSize | %v |\t%v \nRegion | %v | %v |\nRDSInstanceClass | %v | %v|\t", 1, c.ConcourseWorkerCount, "medium", c.ConcourseWorkerSize, "large", c.ConcourseWebSize, "eu-west-3", c.Region, "db.t2.small", c.RDSInstanceClass)
				return isValid, message
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Iaas:         tt.fields.Iaas,
				Project:      tt.fields.Project,
				Namespace:    tt.fields.Namespace,
				BucketName:   tt.fields.BucketName,
				BucketExists: tt.fields.BucketExists,
				BucketError:  tt.fields.BucketError,
				Config:       tt.fields.Config,
			}
			config, newConfigCreated, isDomainUpdated, err := client.LoadOrCreate(tt.args.deployArgs)
			isValid, message := tt.validate(config, newConfigCreated, isDomainUpdated)
			if !isValid {
				t.Errorf("ClientLoadOrCreate()\n %s\n", message)
			}
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("Client.LoadOrCreate()\n expects an error: %v\n received error: %v", tt.wantErr, err)
				return
			}
		})
	}
}
