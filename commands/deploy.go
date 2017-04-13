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
)

type deployCredentials struct {
	PrivateKey string
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
		configDir := "~/.concourse-up"
		keyFile := fmt.Sprintf("%s/concourse-up-%s.pem", configDir, name)
		configPath, err := util.Path(configDir)
		util.CheckErr(err)
		keyPath, err := util.Path(keyFile)
		util.CheckErr(err)

		// Check if config dir exists and create it if it doesn't
		err = util.AssertDirExists(configPath)
		util.CheckErr(err)

		// Check if key file exists and create it if it doesn't
		err = util.AssertFileExists(keyPath)
		util.CheckErr(err)

		// Load key from key file into credentials
		keySource, err := ioutil.ReadFile(keyPath)
		util.CheckErr(err)
		credentials.PrivateKey = string(keySource)

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
		updatedKey := []byte(credentials.PrivateKey)

		err = ioutil.WriteFile(keyPath, updatedKey, 0600)
		util.CheckErr(err)

		return err
	},
}
