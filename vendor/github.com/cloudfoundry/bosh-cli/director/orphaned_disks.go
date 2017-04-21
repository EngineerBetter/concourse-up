package director

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type OrphanedDiskImpl struct {
	client Client

	cid  string
	size uint64

	deploymentName string
	instanceName   string
	azName         string

	orphanedAt time.Time
}

func (d OrphanedDiskImpl) CID() string  { return d.cid }
func (d OrphanedDiskImpl) Size() uint64 { return d.size }

func (d OrphanedDiskImpl) Deployment() Deployment {
	return &DeploymentImpl{client: d.client, name: d.deploymentName}
}

func (d OrphanedDiskImpl) InstanceName() string { return d.instanceName }
func (d OrphanedDiskImpl) AZName() string       { return d.azName }

func (d OrphanedDiskImpl) OrphanedAt() time.Time { return d.orphanedAt }

func (d OrphanedDiskImpl) Delete() error {
	err := d.client.DeleteOrphanedDisk(d.cid)
	if err != nil {
		resps, listErr := d.client.OrphanedDisks()
		if listErr != nil {
			return err
		}

		for _, resp := range resps {
			if resp.CID == d.cid {
				return err
			}
		}
	}

	return nil
}

type OrphanedDiskResp struct {
	CID  string `json:"disk_cid"`
	Size uint64

	DeploymentName string `json:"deployment_name"`
	InstanceName   string `json:"instance_name"`
	AZ             string `json:"az"`

	OrphanedAt string `json:"orphaned_at"` // e.g. "2016-01-09 06:23:25 +0000"
}

func (d DirectorImpl) FindOrphanedDisk(cid string) (OrphanedDisk, error) {
	return OrphanedDiskImpl{client: d.client, cid: cid}, nil
}

func (d DirectorImpl) OrphanedDisks() ([]OrphanedDisk, error) {
	var disks []OrphanedDisk

	resps, err := d.client.OrphanedDisks()
	if err != nil {
		return disks, err
	}

	for _, r := range resps {
		orphanedAt, err := TimeParser{}.Parse(r.OrphanedAt)
		if err != nil {
			return disks, bosherr.WrapErrorf(err, "Converting orphaned at '%s' to time", r.OrphanedAt)
		}

		disk := OrphanedDiskImpl{
			client: d.client,

			cid:  r.CID,
			size: r.Size,

			deploymentName: r.DeploymentName,
			instanceName:   r.InstanceName,
			azName:         r.AZ,

			orphanedAt: orphanedAt.UTC(),
		}

		disks = append(disks, disk)
	}

	return disks, nil
}

func (c Client) OrphanedDisks() ([]OrphanedDiskResp, error) {
	var disks []OrphanedDiskResp

	err := c.clientRequest.Get("/disks", &disks)
	if err != nil {
		return disks, bosherr.WrapErrorf(err, "Finding orphaned disks")
	}

	return disks, nil
}

func (c Client) DeleteOrphanedDisk(cid string) error {
	if len(cid) == 0 {
		return bosherr.Error("Expected non-empty orphaned disk CID")
	}

	path := fmt.Sprintf("/disks/%s", cid)

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting orphaned disk '%s'", cid)
	}

	return nil
}
