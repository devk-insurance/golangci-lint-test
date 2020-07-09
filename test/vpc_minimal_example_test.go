package test

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const (
	keyExpectedName      = "expected_name"
	keyExpectedCIDRBlock = "expected_cidr_block"

	inputName      = "name"
	inputCIDRBlock = "cidr_block"

	outputVpcID        = "vpc_id"
	outputVpcCIDRBlock = "vpc_cidr_block"
)

func TestVPCMinimal(t *testing.T) {
	t.Parallel()

	const workingDir = "../examples/vpc-minimal-example"
	defer test_structure.RunTestStage(t, keyStageCleanup, func() {
		terraform.Destroy(t, test_structure.LoadTerraformOptions(t, workingDir))
	})
	test_structure.RunTestStage(t, keyStageSetup, func() {
		setupVPCMinimal(t, workingDir)
	})
	test_structure.RunTestStage(t, keyStageDeploy, func() {
		terraform.InitAndApply(t, test_structure.LoadTerraformOptions(t, workingDir))
	})
	test_structure.RunTestStage(t, keyStageValidate, func() {
		validateVPCMinimal(t, workingDir)
	})
}

func setupVPCMinimal(t *testing.T, workingDir string) {
	var (
		name      = fmt.Sprintf("testvpc-%s", strings.ToLower(random.UniqueId()))
		cidrBlock = "192.168.100.0/22"
		awsRegion = getAWSRegion(t)
	)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: workingDir,

		Vars: map[string]interface{}{
			inputName:      name,
			inputCIDRBlock: cidrBlock,
		},

		// Environment variables to set when running Terraform
		EnvVars: map[string]string{
			keyAWSDefaultRegion: awsRegion,
		},
	}

	test_structure.SaveString(t, workingDir, keyExpectedCIDRBlock, cidrBlock)
	test_structure.SaveString(t, workingDir, keyExpectedName, name)
	test_structure.SaveTerraformOptions(t, workingDir, terraformOptions)
}

func validateVPCMinimal(t *testing.T, workingDir string) {
	terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
	awsRegion := terraformOptions.EnvVars[keyAWSDefaultRegion]

	cidrBlock := terraform.OutputRequired(t, terraformOptions, outputVpcCIDRBlock)
	expectedCIDRBlock := test_structure.LoadString(t, workingDir, keyExpectedCIDRBlock)
	assert.Exactly(t, expectedCIDRBlock, cidrBlock, "CIDR-Block should stay the same")

	vpcID := terraform.OutputRequired(t, terraformOptions, outputVpcID)
	vpcIDFetched := aws.GetVpcById(t, vpcID, awsRegion)
	expectedVPCName := test_structure.LoadString(t, workingDir, keyExpectedName)
	assert.Exactly(t, expectedVPCName, vpcIDFetched.Name, "VPC name should match")
}
