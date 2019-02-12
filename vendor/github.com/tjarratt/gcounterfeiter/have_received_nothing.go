package gcounterfeiter

import (
	"fmt"

	"github.com/tjarratt/gcounterfeiter/invocations"
)

type haveReceivedNothingMatcher struct {
	expected                invocations.Recorder
	functionWasInvokedCount int
}

func (m *haveReceivedNothingMatcher) Match(expected interface{}) (bool, error) {
	fake, ok := expected.(invocations.Recorder)
	if !ok {
		return false, expectedDoesNotImplementInterfaceError(expected)
	}

	m.expected = fake
	return len(fake.Invocations()) != 0, nil
}

func (m *haveReceivedNothingMatcher) FailureMessage(interface{}) string {
	return fmt.Sprintf("Expected to have received at least one invocation, but it received %d", invocations.CountTotalInvocations(m.expected.Invocations()))
}

func (m *haveReceivedNothingMatcher) NegatedFailureMessage(interface{}) string {
	return fmt.Sprintf("Expected to have received nothing, but there were %d invocations", invocations.CountTotalInvocations(m.expected.Invocations()))
}

func (m *haveReceivedNothingMatcher) With(_ ...interface{}) HaveReceivableMatcher {
	return newUserDoneGoofedMatcher(incorrectHaveReceivedAndWithUsageMessage)
}

func (m *haveReceivedNothingMatcher) AndWith(_ ...interface{}) HaveReceivableMatcher {
	return newUserDoneGoofedMatcher(incorrectHaveReceivedAndWithUsageMessage)
}

const incorrectHaveReceivedAndWithUsageMessage = `Aww shucks.

You done goofed!
You cannot combine HaveReceived() with argument matching.

HaveReceived() on its own means "a test-double received ... some function".
HaveReceived().With(Equal("something") would mean "a test-double received ... some function?, with the argument 'something'", which is probably not what you wanted.

Perhaps you meant HaveReceived("SomeMethodToTest").With("my-special-argument") ?`
