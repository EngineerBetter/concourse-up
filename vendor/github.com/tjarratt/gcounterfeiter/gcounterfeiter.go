package gcounterfeiter

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

func HaveReceived(args ...string) HaveReceivableMatcher {
	switch len(args) {
	case 0:
		return &haveReceivedNothingMatcher{}
	case 1:
		return &haveReceivedMatcher{functionToMatch: args[0]}
	default:
		return &userDoneGoofedMatcher{
			message: fmt.Sprintf("You provided too many arguments. Expected 0 or 1, but you provided %d", len(args)),
		}
	}
}

type HaveReceivableMatcher interface {
	types.GomegaMatcher
	With(...interface{}) HaveReceivableMatcher
	AndWith(...interface{}) HaveReceivableMatcher
}
