package bosh

const concourseStemcellURL = "https://bosh-jenkins-artifacts.s3.amazonaws.com/bosh-stemcell/aws/light-bosh-stemcell-3262.4.1-aws-xen-ubuntu-trusty-go_agent.tgz"

var concourseReleaseURLs = []string{
	"https://bosh.io/d/github.com/concourse/concourse?v=2.7.3",
	"https://bosh.io/d/github.com/cloudfoundry/garden-runc-release?v=1.4.0",
}

func (client *Client) uploadConcourse() error {
	_, err := client.runAuthenticatedBoshCommand(
		"upload-stemcell",
		concourseStemcellURL,
	)
	if err != nil {
		return err
	}

	for _, releaseURL := range concourseReleaseURLs {
		_, err := client.runAuthenticatedBoshCommand(
			"upload-release",
			releaseURL,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
