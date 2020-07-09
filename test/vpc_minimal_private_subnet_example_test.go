package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

const (
	keyExpectedAvailabilityZonesLength = "expected_az_len"
	inputAvailabilityZones             = "availability_zones"
	outputSubnetCIDRBlocks             = "subnet_cidr_blocks"
)

func TestVPCMinimalSubnet(t *testing.T) {
	t.Parallel()

	workingDir := "../examples/vpc-minimal-private-subnet-example"
	defer test_structure.RunTestStage(t, keyStageCleanup, func() {
		terraform.Destroy(t, test_structure.LoadTerraformOptions(t, workingDir))
	})
	test_structure.RunTestStage(t, keyStageSetup, func() {
		setupVPCMinimalSubnet(t, workingDir)
	})
	test_structure.RunTestStage(t, keyStageDeploy, func() {
		terraform.InitAndApply(t, test_structure.LoadTerraformOptions(t, workingDir))
	})
	test_structure.RunTestStage(t, keyStageValidate, func() {
		validateVPCMinimalSubnet(t, workingDir)
	})
}

func setupVPCMinimalSubnet(t *testing.T, workingDir string) {
	var (
		name              = fmt.Sprintf("testvpc-%s", strings.ToLower(random.UniqueId()))
		cidrBlock         = "192.168.100.0/22"
		availabilityZones = []string{
			"eu-central-1a",
			"eu-central-1b",
			"eu-central-1c",
		}
		awsRegion = getAWSRegion(t)
	)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: workingDir,

		Vars: map[string]interface{}{
			inputName:              name,
			inputCIDRBlock:         cidrBlock,
			inputAvailabilityZones: availabilityZones,
		},

		// Environment variables to set when running Terraform
		EnvVars: map[string]string{
			keyAWSDefaultRegion: awsRegion,
		},
	}

	test_structure.SaveInt(t, workingDir, keyExpectedAvailabilityZonesLength, len(availabilityZones))
	test_structure.SaveString(t, workingDir, keyExpectedCIDRBlock, cidrBlock)
	test_structure.SaveTerraformOptions(t, workingDir, terraformOptions)
}

func validateVPCMinimalSubnet(t *testing.T, workingDir string) {
	terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
	awsRegion := terraformOptions.EnvVars[keyAWSDefaultRegion]

	cidrBlock := terraform.OutputRequired(t, terraformOptions, outputVpcCIDRBlock)
	expectedCIDRBlock := test_structure.LoadString(t, workingDir, keyExpectedCIDRBlock)
	assert.Exactly(t, expectedCIDRBlock, cidrBlock, "CIDR-Block should stay the same")

	subnetCIDRBlocks := terraform.OutputList(t, terraformOptions, outputSubnetCIDRBlocks)
	expectedSubnetBlockLength := test_structure.LoadInt(t, workingDir, keyExpectedAvailabilityZonesLength)
	assert.Len(t, subnetCIDRBlocks, expectedSubnetBlockLength, "should have the same amount of subnets than availability zones")

	// subnets
	vpcID := terraform.OutputRequired(t, terraformOptions, outputVpcID)
	subnets := aws.GetSubnetsForVpc(t, vpcID, awsRegion)
	assert.Len(t, subnets, expectedSubnetBlockLength, "should have the same amount of subnets via terraform output or AWS api")
	for _, subnetID := range subnets {
		assert.False(t, aws.IsPublicSubnet(t, subnetID.Id, awsRegion))
	}
}
