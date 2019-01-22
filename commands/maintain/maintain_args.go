package maintain

import (
	"fmt"

	cli "gopkg.in/urfave/cli.v1"
)

// Args are arguments passed to the info command
type Args struct {
	AWSRegion          string
	AWSRegionIsSet     bool
	RenewNatsCert      bool
	RenewNatsCertIsSet bool
	Namespace          string
	NamespaceIsSet     bool
	IAAS               string
	Stage              int
	StageIsSet         bool
}

//MarkSetFlags is marking which info Args have been set
func (a *Args) MarkSetFlags(c FlagSetChecker) error {
	for _, f := range c.FlagNames() {
		if c.IsSet(f) {
			switch f {
			case "region":
				a.AWSRegionIsSet = true
			case "namespace":
				a.NamespaceIsSet = true
			case "renew-nats-cert":
				a.RenewNatsCertIsSet = true
			case "stage":
				a.StageIsSet = true
			case "iaas":
				//do nothing
			default:
				return fmt.Errorf("flag %q is not supported by maintain flags", f)
			}
		}
	}
	return nil
}

// FlagSetChecker allows us to find out if flags were set, adn what the names of all flags are
type FlagSetChecker interface {
	IsSet(name string) bool
	FlagNames() (names []string)
}

// ContextWrapper wraps a CLI context for testing
type ContextWrapper struct {
	c *cli.Context
}

// IsSet tells you if a user provided a flag
func (t *ContextWrapper) IsSet(name string) bool {
	return t.c.IsSet(name)
}

// FlagNames lists all flags it's possible for a user to provide
func (t *ContextWrapper) FlagNames() (names []string) {
	return t.c.FlagNames()
}
