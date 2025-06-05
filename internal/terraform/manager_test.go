package terraform

import (
	"os"
	"testing"

	"github.com/sorinlg/tf-manage2/internal/config"
	"github.com/sorinlg/tf-manage2/internal/framework"
)

func TestDetectExecMode(t *testing.T) {
	// Create a basic config for testing
	cfg := &config.Config{
		RepoName: "test-repo",
	}
	manager := NewManager(cfg)

	expectUnattended := framework.AddEmphasisRed("unattended")
	expectOperator := framework.AddEmphasisGreen("operator")

	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "Default operator mode",
			envVars:  map[string]string{},
			expected: expectOperator,
		},
		{
			name: "Manual override",
			envVars: map[string]string{
				"TF_EXEC_MODE_OVERRIDE": "1",
			},
			expected: expectUnattended,
		},
		{
			name: "GitHub Actions",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			expected: expectUnattended,
		},
		{
			name: "GitLab CI",
			envVars: map[string]string{
				"GITLAB_CI": "true",
			},
			expected: expectUnattended,
		},
		{
			name: "CircleCI",
			envVars: map[string]string{
				"CIRCLECI": "true",
			},
			expected: expectUnattended,
		},
		{
			name: "Travis CI",
			envVars: map[string]string{
				"TRAVIS": "true",
			},
			expected: expectUnattended,
		},
		{
			name: "Azure DevOps",
			envVars: map[string]string{
				"TF_BUILD": "True",
			},
			expected: expectUnattended,
		},
		{
			name: "Jenkins URL",
			envVars: map[string]string{
				"JENKINS_URL": "http://jenkins.example.com",
			},
			expected: expectUnattended,
		},
		{
			name: "Jenkins BUILD_NUMBER",
			envVars: map[string]string{
				"BUILD_NUMBER": "123",
			},
			expected: expectUnattended,
		},
		{
			name: "Legacy Jenkins user",
			envVars: map[string]string{
				"USER": "jenkins",
			},
			expected: expectUnattended,
		},
		{
			name: "Bamboo",
			envVars: map[string]string{
				"bamboo_buildKey": "TEST-PLAN-123",
			},
			expected: expectUnattended,
		},
		{
			name: "TeamCity",
			envVars: map[string]string{
				"TEAMCITY_VERSION": "2021.1",
			},
			expected: expectUnattended,
		},
		{
			name: "Buildkite",
			envVars: map[string]string{
				"BUILDKITE": "true",
			},
			expected: expectUnattended,
		},
		{
			name: "Drone CI",
			envVars: map[string]string{
				"DRONE": "true",
			},
			expected: expectUnattended,
		},
		{
			name: "AWS CodeBuild",
			envVars: map[string]string{
				"CODEBUILD_BUILD_ID": "test-build-123",
			},
			expected: expectUnattended,
		},
		{
			name: "Generic CI (true)",
			envVars: map[string]string{
				"CI": "true",
			},
			expected: expectUnattended,
		},
		{
			name: "Generic CI (1)",
			envVars: map[string]string{
				"CI": "1",
			},
			expected: expectUnattended,
		},
		{
			name: "Override takes precedence over CI",
			envVars: map[string]string{
				"TF_EXEC_MODE_OVERRIDE": "1",
				"GITHUB_ACTIONS":        "true",
			},
			expected: expectUnattended,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			clearCIEnvVars()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Test the detection
			result := manager.detectExecMode()
			if result != tt.expected {
				t.Errorf("detectExecMode() = %v, want %v", result, tt.expected)
			}

			// Clean up environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestIsRunningInCI(t *testing.T) {
	cfg := &config.Config{
		RepoName: "test-repo",
	}
	manager := NewManager(cfg)

	// Test that multiple CI variables work
	clearCIEnvVars()
	os.Setenv("GITHUB_ACTIONS", "true")
	os.Setenv("CI", "true")

	if !manager.isRunningInCI() {
		t.Error("Expected CI detection to return true with multiple CI variables set")
	}

	// Clean up
	clearCIEnvVars()

	// Test that no CI variables returns false
	if manager.isRunningInCI() {
		t.Error("Expected CI detection to return false with no CI variables set")
	}
}

// clearCIEnvVars removes all CI-related environment variables for clean testing
func clearCIEnvVars() {
	ciVars := []string{
		"TF_EXEC_MODE_OVERRIDE",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"CIRCLECI",
		"TRAVIS",
		"TF_BUILD",
		"JENKINS_URL",
		"BUILD_NUMBER",
		"USER",
		"bamboo_buildKey",
		"TEAMCITY_VERSION",
		"BUILDKITE",
		"DRONE",
		"CODEBUILD_BUILD_ID",
		"CI",
	}

	for _, envVar := range ciVars {
		os.Unsetenv(envVar)
	}
}
