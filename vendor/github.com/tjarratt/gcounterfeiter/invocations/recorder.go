package invocations

type Recorder interface {
	Invocations() map[string][][]interface{}
}
