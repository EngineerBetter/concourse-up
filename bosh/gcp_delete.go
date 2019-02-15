package bosh

import (
	"net"

	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
	"github.com/EngineerBetter/concourse-up/bosh/internal/gcp"
	"github.com/apparentlymart/go-cidr/cidr"
)

// Delete deletes a bosh director
func (client *GCPClient) Delete(stateFileBytes []byte) ([]byte, error) {
	if err := client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		false,
		"--deployment",
		concourseDeploymentName,
		"delete-deployment",
		"--force",
	); err != nil {
		return nil, err
	}

	//TODO(px): pull up this so that we use aws.Store
	store := temporaryStore{
		"state.json": stateFileBytes,
	}
	bosh, err := boshenv.New()
	if err != nil {
		return store["state.json"], err
	}

	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return store["state.json"], err
	}

	publicCIDR := client.config.PublicCIDR
	_, pubCIDR, err := net.ParseCIDR(publicCIDR)
	if err != nil {
		return store["state.json"], err
	}
	internalGateway, err := cidr.Host(pubCIDR, 1)
	if err != nil {
		return store["state.json"], err
	}
	directorInternalIP, err := cidr.Host(pubCIDR, 6)
	if err != nil {
		return store["state.json"], err
	}
	credentialsPath, err := client.provider.Attr("credentials_path")
	if err != nil {
		return store["state.json"], err
	}
	network, err := client.outputs.Get("Network")
	if err != nil {
		return store["state.json"], err
	}
	publicSubnetwork, err := client.outputs.Get("PublicSubnetworkName")
	if err != nil {
		return store["state.json"], err
	}
	privateSubnetwork, err := client.outputs.Get("PrivateSubnetworkName")
	if err != nil {
		return store["state.json"], err
	}
	project, err := client.provider.Attr("project")
	if err != nil {
		return store["state.json"], err
	}

	err = bosh.DeleteEnv(store, gcp.Environment{
		DirectorName:       "bosh",
		ExternalIP:         directorPublicIP,
		GcpCredentialsJSON: credentialsPath,
		InternalCIDR:       client.config.PublicCIDR,
		InternalGW:         internalGateway.String(),
		InternalIP:         directorInternalIP.String(),
		Network:            network,
		PrivateSubnetwork:  privateSubnetwork,
		ProjectID:          project,
		PublicKey:          client.config.PublicKey,
		PublicSubnetwork:   publicSubnetwork,
		Spot:               client.config.Spot,
		Zone:               client.provider.Zone(""),
	}, client.config.DirectorPassword, client.config.DirectorCert, client.config.DirectorKey, client.config.DirectorCACert, nil)
	return store["state.json"], err
}
