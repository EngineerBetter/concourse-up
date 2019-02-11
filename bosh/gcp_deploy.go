package bosh

import (
	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
	"github.com/EngineerBetter/concourse-up/bosh/internal/gcp"
)

// Deploy deploys a new Bosh director or converges an existing deployment
// Returns new contents of bosh state file
func (client *GCPClient) Deploy(state, creds []byte, detach bool) (newState, newCreds []byte, err error) {
	bosh, err := boshenv.New(boshenv.DownloadBOSH())
	if err != nil {
		return state, creds, err
	}

	state, creds, err = client.createEnv(bosh, state, creds, "")
	if err != nil {
		return state, creds, err
	}

	if err = client.updateCloudConfig(bosh); err != nil {
		return state, creds, err
	}
	if err = client.uploadConcourseStemcell(bosh); err != nil {
		return state, creds, err
	}
	if err = client.createDefaultDatabases(); err != nil {
		return state, creds, err
	}

	creds, err = client.deployConcourse(creds, detach)
	if err != nil {
		return state, creds, err
	}

	return state, creds, err
}

// CreateEnv exposes bosh create-env functionality
func (client *GCPClient) CreateEnv(state, creds []byte, customOps string) (newState, newCreds []byte, err error) {
	bosh, err := boshenv.New(boshenv.DownloadBOSH())
	if err != nil {
		return state, creds, err
	}
	return client.createEnv(bosh, state, creds, customOps)
}

// Recreate exposes BOSH recreate
func (client *GCPClient) Recreate() error {
	bosh, err := boshenv.New(boshenv.DownloadBOSH())
	if err != nil {
		return err
	}
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.Recreate(gcp.Environment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
}

func (client *GCPClient) createEnv(bosh *boshenv.BOSHCLI, state, creds []byte, customOps string) (newState, newCreds []byte, err error) {
	tags, err := splitTags(client.config.Tags)
	if err != nil {
		return state, creds, err
	}
	tags["concourse-up-project"] = client.config.Project
	tags["concourse-up-component"] = "concourse"
	//TODO(px): pull up this so that we use aws.Store
	store := temporaryStore{
		"vars.yaml":  creds,
		"state.json": state,
	}

	network, err1 := client.outputs.Get("Network")
	if err1 != nil {
		return state, creds, err1
	}
	publicSubnetwork, err1 := client.outputs.Get("PublicSubnetworkName")
	if err1 != nil {
		return state, creds, err1
	}
	privateSubnetwork, err1 := client.outputs.Get("PrivateSubnetworkName")
	if err1 != nil {
		return state, creds, err1
	}
	directorPublicIP, err1 := client.outputs.Get("DirectorPublicIP")
	if err1 != nil {
		return state, creds, err1
	}
	project, err1 := client.provider.Attr("project")
	if err1 != nil {
		return state, creds, err1
	}
	// credentialsPath := "~/Downloads/concourse-up-6f707de4d7bd.json"
	credentialsPath, err1 := client.provider.Attr("credentials_path")
	if err1 != nil {
		return state, creds, err1
	}
	err1 = bosh.CreateEnv(store, gcp.Environment{
		InternalCIDR:       "10.0.0.0/24",
		InternalGW:         "10.0.0.1",
		InternalIP:         "10.0.0.6",
		DirectorName:       "bosh",
		Zone:               client.provider.Zone(""),
		Network:            network,
		PublicSubnetwork:   publicSubnetwork,
		PrivateSubnetwork:  privateSubnetwork,
		Tags:               "[internal]",
		ProjectID:          project,
		GcpCredentialsJSON: credentialsPath,
		ExternalIP:         directorPublicIP,
		Spot:               client.config.Spot,
		PublicKey:          client.config.PublicKey,
		CustomOperations:   customOps,
	}, client.config.DirectorPassword, client.config.DirectorCert, client.config.DirectorKey, client.config.DirectorCACert, tags)
	if err1 != nil {
		return store["state.json"], store["vars.yaml"], err1
	}
	return store["state.json"], store["vars.yaml"], err
}

// Locks implements locks for GCP client
func (client *GCPClient) Locks() ([]byte, error) {
	bosh, err := boshenv.New(boshenv.DownloadBOSH())
	if err != nil {
		return []byte{}, err
	}
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}
	return bosh.Locks(gcp.Environment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)

}

func (client *GCPClient) updateCloudConfig(bosh *boshenv.BOSHCLI) error {

	privateSubnetwork, err := client.outputs.Get("PrivateSubnetworkName")
	if err != nil {
		return err
	}
	publicSubnetwork, err := client.outputs.Get("PublicSubnetworkName")
	if err != nil {
		return err
	}
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	network, err := client.outputs.Get("Network")
	if err != nil {
		return err
	}
	zone := client.provider.Zone("")
	return bosh.UpdateCloudConfig(gcp.Environment{
		Spot:              client.config.Spot,
		PublicSubnetwork:  publicSubnetwork,
		PrivateSubnetwork: privateSubnetwork,
		Zone:              zone,
		Network:           network,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
}
func (client *GCPClient) uploadConcourseStemcell(bosh *boshenv.BOSHCLI) error {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.UploadConcourseStemcell(gcp.Environment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
}

func (client *GCPClient) saveStateFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(StateFilename), nil
	}

	return client.director.SaveFileToWorkingDir(StateFilename, bytes)
}

func (client *GCPClient) saveCredsFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(CredsFilename), nil
	}

	return client.director.SaveFileToWorkingDir(CredsFilename, bytes)
}
