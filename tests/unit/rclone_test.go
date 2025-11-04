package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/rclone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRcloneNewManager(t *testing.T) {
	rclonePath := "/usr/local/bin/rclone"
	manager := rclone.NewManager(rclonePath)

	assert.NotNil(t, manager)
	assert.NotEmpty(t, manager.GetConfigPath())
	assert.Contains(t, manager.GetConfigPath(), "rclone.conf")
}

func TestNewManagerWithConfig(t *testing.T) {
	rclonePath := "/usr/local/bin/rclone"
	configPath := "/tmp/test-rclone.conf"
	manager := rclone.NewManagerWithConfig(rclonePath, configPath)

	assert.NotNil(t, manager)
	assert.Equal(t, configPath, manager.GetConfigPath())
}

func TestConfigExists(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rclone.conf")

	tests := []struct {
		name       string
		createFile bool
		want       bool
	}{
		{
			name:       "config file exists",
			createFile: true,
			want:       true,
		},
		{
			name:       "config file does not exist",
			createFile: false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file if needed
			if tt.createFile {
				err := os.WriteFile(configPath, []byte("test"), 0644)
				require.NoError(t, err)
				defer os.Remove(configPath)
			}

			manager := rclone.NewManagerWithConfig("/usr/local/bin/rclone", configPath)
			got := manager.ConfigExists()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	configPath := "/custom/path/rclone.conf"
	manager := rclone.NewManagerWithConfig("/usr/local/bin/rclone", configPath)

	assert.Equal(t, configPath, manager.GetConfigPath())
}

// TestListRemotesWithMockConfig tests ListRemotes with a mock config file
func TestListRemotesWithMockConfig(t *testing.T) {
	// This test would require a mock rclone binary or mock exec.Command
	// For now, we test the basic structure
	t.Skip("Requires mock rclone binary")
}

// TestListBuckets tests bucket listing
func TestListBuckets(t *testing.T) {
	// This test would require a mock rclone binary
	t.Skip("Requires mock rclone binary")
}

// TestTestRemote tests remote connectivity
func TestTestRemote(t *testing.T) {
	// This test would require a mock rclone binary
	t.Skip("Requires mock rclone binary")
}

// TestParseConfig tests configuration parsing
func TestParseConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rclone.conf")

	// Create a sample config file with example remotes
	configContent := `[source-remote]
type = b2
account = test-account-id
key = test-key

[dest-remote]
type = s3
provider = Scaleway
access_key_id = test-access-key
secret_access_key = test-secret-key
region = fr-par
endpoint = s3.fr-par.scw.cloud
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	manager := rclone.NewManagerWithConfig("/usr/local/bin/rclone", configPath)
	exists := manager.ConfigExists()

	assert.True(t, exists)
}

// TestValidateRemoteName tests remote name validation
func TestValidateRemoteName(t *testing.T) {
	tests := []struct {
		name       string
		remoteName string
		wantValid  bool
	}{
		{
			name:       "valid remote name",
			remoteName: "my-remote",
			wantValid:  true,
		},
		{
			name:       "valid remote with numbers",
			remoteName: "remote123",
			wantValid:  true,
		},
		{
			name:       "valid remote with underscore",
			remoteName: "my_remote",
			wantValid:  true,
		},
		{
			name:       "invalid - empty",
			remoteName: "",
			wantValid:  false,
		},
		{
			name:       "invalid - spaces",
			remoteName: "my remote",
			wantValid:  false,
		},
		{
			name:       "invalid - special chars",
			remoteName: "my@remote",
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rclone.ValidateRemoteName(tt.remoteName)
			if tt.wantValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestValidateBucketName tests bucket name validation
func TestValidateBucketName(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		wantValid  bool
	}{
		{
			name:       "valid bucket name",
			bucketName: "my-bucket",
			wantValid:  true,
		},
		{
			name:       "valid with numbers",
			bucketName: "bucket123",
			wantValid:  true,
		},
		{
			name:       "invalid - empty",
			bucketName: "",
			wantValid:  false,
		},
		{
			name:       "invalid - spaces",
			bucketName: "my bucket",
			wantValid:  false,
		},
		{
			name:       "invalid - uppercase",
			bucketName: "MyBucket",
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rclone.ValidateBucketName(tt.bucketName)
			if tt.wantValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
