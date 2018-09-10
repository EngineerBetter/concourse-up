package config

// InfoArgs are arguments passed to the info command
type InfoArgs struct {
	AWSRegion string
	JSON      bool
	IAAS      string
	Env       bool
	Namespace string
}
