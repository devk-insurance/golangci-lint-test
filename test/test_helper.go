package test

import (
	"github.com/gruntwork-io/terratest/modules/aws"
	"testing"
)

const (
	// Contains various constants for test-stages
	keyStageCleanup  = "cleanup"
	keyStageDeploy   = "deploy"
	keyStageSetup    = "setup"
	keyStageValidate = "validate"
)

const (
	keyAWSDefaultRegion = "AWS_DEFAULT_REGION"
)

// getAWSRegion returns a valid AWS region which can be used to run tests.
func getAWSRegion(t *testing.T) string {
	return aws.GetRandomStableRegion(t, []string{"eu-central-1"}, []string{})
}
