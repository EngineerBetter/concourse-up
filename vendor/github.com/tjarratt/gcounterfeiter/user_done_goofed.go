package gcounterfeiter

import "errors"

type userDoneGoofedMatcher struct {
	message string
}

func newUserDoneGoofedMatcher(message string) *userDoneGoofedMatcher {
	return &userDoneGoofedMatcher{message: message}
}

func (m *userDoneGoofedMatcher) Match(interface{}) (bool, error) {
	return false, errors.New(m.message)
}

func (m *userDoneGoofedMatcher) FailureMessage(interface{}) string {
	return ""
}

func (m *userDoneGoofedMatcher) NegatedFailureMessage(interface{}) string {
	return ""
}

func (m *userDoneGoofedMatcher) With(_ ...interface{}) HaveReceivableMatcher {
	return m
}

func (m *userDoneGoofedMatcher) AndWith(_ ...interface{}) HaveReceivableMatcher {
	return m
}
