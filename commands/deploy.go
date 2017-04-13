package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"bitbucket.org/engineerbetter/concourse-up/aws"
	"bitbucket.org/engineerbetter/concourse-up/util"

	"gopkg.in/urfave/cli.v1"
	yaml "gopkg.in/yaml.v2"
)

type deployCredentials struct {
	PrivateKey string `yaml:"private_key"`
}

var credentials = deployCredentials{}

var awsRegion string

var deployFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Value:       "eu-west-1",
		Usage:       "AWS region",
		Destination: &awsRegion,
	},
}

var deploy = cli.Command{
	Name:      "deploy",
	Aliases:   []string{"d"},
	Usage:     "Deploys or updates a Concourse",
	ArgsUsage: "<name>",
	Flags:     deployFlags,
	Action: func(c *cli.Context) error {

		if len(os.Args) < 3 {
			fmt.Println("Usage is `concourse-up deploy <name>`")
			return nil
		}
		name := os.Args[2]
		configFile := fmt.Sprintf("~/.concourse-up-%s", name)
		path, err := util.Path(configFile)
		util.CheckErr(err)

		// Check if config file exists and create it if it doesn't
		err = util.AssertFileExists(path)
		util.CheckErr(err)

		// Extract contents of config file
		source, err := ioutil.ReadFile(path)
		util.CheckErr(err)

		// Unmarshal yaml from config file into credentials
		err = yaml.Unmarshal(source, &credentials)
		util.CheckErr(err)

		fmt.Printf("Checking for keypair %s-bosh and generating if missing\n", name)
		awsSession, err := aws.CreateSession()
		if err != nil {
			return err
		}
		ec2Client := ec2.New(awsSession, &awssdk.Config{Region: &awsRegion})
		updatedPrivateKey, err := aws.AssertAWSKeyPairExists(
			ec2Client,
			fmt.Sprintf("%s-bosh", name),
			credentials.PrivateKey)
		if err != nil {
			return err
		}
		credentials.PrivateKey = updatedPrivateKey
		updatedSource, err := yaml.Marshal(&credentials)
		util.CheckErr(err)

		err = ioutil.WriteFile(path, updatedSource, 0600)
		util.CheckErr(err)

		return err
	},
}
