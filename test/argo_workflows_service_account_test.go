package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

// Test using test_structure, which allows sections of the tests to be skipped
// if a `SKIP_stage_name=true` environment variable is set.  This allows for
// selectively skipping long running test steps like terraform apply/destroy.
func TestArgoWorkflowsServiceAccount(t *testing.T) {
	t.Parallel()

	workingDir := "../examples/argo-workflows-service-account"

	expectedNamespace := "argo-workflows-system"
	expectedRoleName := "argo-workflows-workflow"
	expectedRoleBindingName := "argo-workflows-workflow"

	terraformOptions := &terraform.Options{
		TerraformDir: workingDir,
	}
	test_structure.SaveTerraformOptions(t, workingDir, terraformOptions)

	// Destroy
	defer test_structure.RunTestStage(t, "destroy", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.Destroy(t, terraformOptions)
	})

	// Apply
	test_structure.RunTestStage(t, "apply", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.InitAndApply(t, terraformOptions)

		namespace := terraform.OutputRequired(t, terraformOptions, "helm_namespace")

		assert.Equal(t, expectedNamespace, namespace)
	})

	// Check
	test_structure.RunTestStage(t, "k8s_objects", func() {
		kubectlOptions := k8s.NewKubectlOptions("", "", expectedNamespace)

		// Check the role
		role := k8s.GetRole(t, kubectlOptions, expectedRoleName)

		// Check the rules
		assert.Equal(t, 2, len(role.Rules))

		// Check the RoleBinding
		roleBinding := getRoleBinding(t, kubectlOptions, expectedNamespace, expectedRoleBindingName)

		assert.Equal(t, "Role", roleBinding.RoleRef.Kind)
		assert.Equal(t, expectedRoleName, roleBinding.RoleRef.Name)
		assert.Equal(t, "rbac.authorization.k8s.io", roleBinding.RoleRef.APIGroup)

		assert.Equal(t, 1, len(roleBinding.Subjects))
		assert.Equal(t, "ServiceAccount", roleBinding.Subjects[0].Kind)
		assert.Equal(t, "argo-workflow", roleBinding.Subjects[0].Name)
	})
}
