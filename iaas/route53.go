package iaas

import "github.com/aws/aws-sdk-go/service/route53"

// Route53 is a substitution for the non-existing interface in the AWS SDK
// The interface is used as an abstraction layer
// to enable mocking of calls to Route53 during testing
//go:generate counterfeiter . Route53
type Route53 interface {
	ListHostedZonesPages(input *route53.ListHostedZonesInput, callback func(output *route53.ListHostedZonesOutput, lastPage bool) (shouldContinue bool)) error
}
