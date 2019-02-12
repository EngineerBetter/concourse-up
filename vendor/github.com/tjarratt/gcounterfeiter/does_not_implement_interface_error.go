package gcounterfeiter

import "fmt"

func expectedDoesNotImplementInterfaceError(expected interface{}) error {
	return fmt.Errorf(`Aww shucks...
You gave us a type we couldn't handle. Perhaps it's not a counterfeiter fake?

Expected value '%v' does not conform to the 'InvocationRecorder' interface.
See HaveReceived() matcher does for more information.`, expected)
}
