package gcounterfeiter

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"github.com/tjarratt/gcounterfeiter/invocations"
)

type argumentVerifyingMatcher struct {
	functionToMatch string
	baseMatcher     types.GomegaMatcher
	argMatchers     []types.GomegaMatcher

	expected invocations.Recorder

	wasNotInvoked        bool
	failedArgIndex       int
	failedMatcherMessage string
}

func NewArgumentVerifyingMatcher(baseMatcher types.GomegaMatcher, functionToMatch string, argMatchers ...types.GomegaMatcher) *argumentVerifyingMatcher {
	return &argumentVerifyingMatcher{
		baseMatcher:     baseMatcher,
		functionToMatch: functionToMatch,
		argMatchers:     argMatchers,
	}
}

func (m *argumentVerifyingMatcher) Match(expected interface{}) (bool, error) {
	ok, err := m.baseMatcher.Match(expected)
	if !ok || err != nil {
		m.wasNotInvoked = true
		m.failedMatcherMessage = m.baseMatcher.FailureMessage(expected)
		return ok, err
	}

	fake, ok := expected.(invocations.Recorder)
	if !ok {
		return false, expectedDoesNotImplementInterfaceError(expected)
	}

	m.expected = fake

	invocations := fake.Invocations()[m.functionToMatch]
	if len(invocations) == 0 {
		m.wasNotInvoked = true
		return false, nil
	}

	var matched bool
	for _, invocation := range invocations {
		if len(invocation) > len(m.argMatchers) {
			return false, fmt.Errorf("Too few arguments provided for '%s'. Expected %d but received %d", m.functionToMatch, len(invocation), len(m.argMatchers))
		}
		if len(invocation) < len(m.argMatchers) {
			return false, fmt.Errorf("Too many arguments provided for '%s'. Expected %d but received %d", m.functionToMatch, len(invocation), len(m.argMatchers))
		}

		matched = true

		for i, arg := range invocation {
			matcher := m.argMatchers[i]

			ok, err := matcher.Match(arg)
			matched = matched && ok
			if err != nil || !ok {
				m.failedArgIndex = i + 1
				m.failedMatcherMessage = matcher.FailureMessage(arg)
			}
		}

		if matched {
			break
		}
	}

	return matched, nil
}

func (m *argumentVerifyingMatcher) FailureMessage(interface{}) string {
	if m.wasNotInvoked {
		return m.failedMatcherMessage
	}

	return fmt.Sprintf(`
Expected to receive '%s' (and it did!) but the %d argument failed to match:

'%s'
`, m.functionToMatch, m.failedArgIndex, m.failedMatcherMessage)
}

func (m *argumentVerifyingMatcher) NegatedFailureMessage(interface{}) string {
	return fmt.Sprintf("Expected to not receive '%s' (with exact argument matching)", m.functionToMatch)
}

func (m *argumentVerifyingMatcher) With(matchersOrValues ...interface{}) HaveReceivableMatcher {
	for _, matcherOrValue := range matchersOrValues {
		argumentMatcher := matcherOrWrapValueWithEqual(matcherOrValue)
		m.argMatchers = append(m.argMatchers, argumentMatcher)
	}
	return m
}

func (m *argumentVerifyingMatcher) AndWith(matchersOrValue ...interface{}) HaveReceivableMatcher {
	return m.With(matchersOrValue...)
}
