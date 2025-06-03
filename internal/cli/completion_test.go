package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sorinlg/tf-manage2/internal/config"
)

// TestCompletion tests the completion functionality
func TestCompletion(t *testing.T) {
	// Create a temporary test directory structure
	tmpDir := t.TempDir()

	// Create test directory structure
	testDirs := []string{
		"terraform/environments/product1/dev/sample_module",
		"terraform/environments/product1/staging/sample_module",
		"terraform/environments/product2/prod/another_module",
		"terraform/modules/sample_module",
		"terraform/modules/another_module",
	}

	for _, dir := range testDirs {
		err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test config files
	testFiles := []string{
		"terraform/environments/product1/dev/sample_module/instance_x.tfvars",
		"terraform/environments/product1/dev/sample_module/instance_y.tfvars",
		"terraform/environments/product1/staging/sample_module/staging_instance.tfvars",
		"terraform/environments/product2/prod/another_module/prod_instance.tfvars",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tmpDir, file)
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Create .tfm.conf
	confContent := `#!/bin/bash
export __tfm_repo_name='test-repo'
export __tfm_env_rel_path='terraform/environments'
export __tfm_module_rel_path='terraform/modules'
`
	confPath := filepath.Join(tmpDir, ".tfm.conf")
	err := os.WriteFile(confPath, []byte(confContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .tfm.conf: %v", err)
	}

	// Create .git directory to simulate git repo
	gitDir := filepath.Join(tmpDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Change to test directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		err := os.Chdir(originalDir)
		if err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	completion := NewCompletion(cfg)

	// Test products completion
	t.Run("SuggestProducts", func(t *testing.T) {
		// Capture stdout
		output := captureOutput(t, func() {
			err := completion.SuggestProducts()
			if err != nil {
				t.Errorf("SuggestProducts failed: %v", err)
			}
		})

		expected := []string{"product1", "product2"}
		for _, exp := range expected {
			if !strings.Contains(output, exp) {
				t.Errorf("Expected product %s not found in output: %s", exp, output)
			}
		}
	})

	// Test modules completion
	t.Run("SuggestModules", func(t *testing.T) {
		output := captureOutput(t, func() {
			err := completion.SuggestModules()
			if err != nil {
				t.Errorf("SuggestModules failed: %v", err)
			}
		})

		expected := []string{"sample_module", "another_module"}
		for _, exp := range expected {
			if !strings.Contains(output, exp) {
				t.Errorf("Expected module %s not found in output: %s", exp, output)
			}
		}
	})

	// Test environments completion
	t.Run("SuggestEnvironments", func(t *testing.T) {
		output := captureOutput(t, func() {
			err := completion.SuggestEnvironments("product1", "sample_module")
			if err != nil {
				t.Errorf("SuggestEnvironments failed: %v", err)
			}
		})

		expected := []string{"dev", "staging"}
		for _, exp := range expected {
			if !strings.Contains(output, exp) {
				t.Errorf("Expected environment %s not found in output: %s", exp, output)
			}
		}

		// Should not contain prod since product1 doesn't have prod/sample_module
		if strings.Contains(output, "prod") {
			t.Errorf("Unexpected environment 'prod' found in output: %s", output)
		}
	})

	// Test configs completion
	t.Run("SuggestConfigs", func(t *testing.T) {
		output := captureOutput(t, func() {
			err := completion.SuggestConfigs("product1", "dev", "sample_module")
			if err != nil {
				t.Errorf("SuggestConfigs failed: %v", err)
			}
		})

		expected := []string{"instance_x", "instance_y"}
		for _, exp := range expected {
			if !strings.Contains(output, exp) {
				t.Errorf("Expected config %s not found in output: %s", exp, output)
			}
		}
	})

	// Test actions completion
	t.Run("SuggestActions", func(t *testing.T) {
		output := captureOutput(t, func() {
			err := completion.SuggestActions()
			if err != nil {
				t.Errorf("SuggestActions failed: %v", err)
			}
		})

		expected := []string{"init", "plan", "apply", "destroy"}
		for _, exp := range expected {
			if !strings.Contains(output, exp) {
				t.Errorf("Expected action %s not found in output: %s", exp, output)
			}
		}
	})
}

// captureOutput captures stdout during function execution
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	// Create a pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Save original stdout
	originalStdout := os.Stdout
	defer func() {
		os.Stdout = originalStdout
	}()

	// Replace stdout with pipe writer
	os.Stdout = w

	// Channel to capture output and errors
	outputChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start goroutine to read from pipe
	go func() {
		defer r.Close()

		var output strings.Builder
		buf := make([]byte, 1024)

		for {
			n, err := r.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				errChan <- err
				return
			}
		}
		outputChan <- output.String()
	}()

	// Execute function
	fn()

	// Close writer to signal EOF
	w.Close()
	os.Stdout = originalStdout

	// Wait for output or error with timeout
	select {
	case output := <-outputChan:
		return output
	case err := <-errChan:
		t.Fatalf("Error reading captured output: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for captured output")
	}

	return ""
}
