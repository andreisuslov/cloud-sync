package unit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreisuslov/cloud-sync/internal/lockfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockfileNewManager(t *testing.T) {
	logDir := "/tmp/logs"
	manager := lockfile.NewManager(logDir)

	assert.NotNil(t, manager)
	assert.Contains(t, manager.GetPath(), "rclone_backup.lock")
}

func TestLockfileNewManagerWithPath(t *testing.T) {
	lockfilePath := "/custom/path/backup.lock"
	manager := lockfile.NewManagerWithPath(lockfilePath)

	assert.NotNil(t, manager)
	assert.Equal(t, lockfilePath, manager.GetPath())
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Initially should not exist
	assert.False(t, manager.Exists())

	// Create the file
	err := os.WriteFile(lockfilePath, []byte("test"), 0644)
	require.NoError(t, err)
	defer os.Remove(lockfilePath)

	// Now should exist
	assert.True(t, manager.Exists())
}

func TestCreate(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Create lockfile
	err := manager.Create()
	require.NoError(t, err)
	defer os.Remove(lockfilePath)

	// Verify it exists
	assert.True(t, manager.Exists())

	// Verify content
	content, err := os.ReadFile(lockfilePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Created:")
}

func TestCreateAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Create lockfile
	err := manager.Create()
	require.NoError(t, err)
	defer os.Remove(lockfilePath)

	// Try to create again - should fail
	err = manager.Create()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Create lockfile
	err := manager.Create()
	require.NoError(t, err)

	// Verify it exists
	assert.True(t, manager.Exists())

	// Remove it
	err = manager.Remove()
	require.NoError(t, err)

	// Verify it's gone
	assert.False(t, manager.Exists())
}

func TestRemoveNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "nonexistent.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Remove non-existent lockfile - should not error
	err := manager.Remove()
	assert.NoError(t, err)
}

func TestGetAge(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Create lockfile
	err := manager.Create()
	require.NoError(t, err)
	defer os.Remove(lockfilePath)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Get age
	age, err := manager.GetAge()
	require.NoError(t, err)
	assert.True(t, age >= 100*time.Millisecond, "Age should be at least 100ms")
	assert.True(t, age < 5*time.Second, "Age should be less than 5 seconds")
}

func TestGetAgeNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "nonexistent.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Get age of non-existent lockfile
	_, err := manager.GetAge()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestIsStale(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Create lockfile
	err := manager.Create()
	require.NoError(t, err)
	defer os.Remove(lockfilePath)

	tests := []struct {
		name   string
		maxAge time.Duration
		want   bool
	}{
		{
			name:   "very short maxAge - should be stale",
			maxAge: 1 * time.Nanosecond,
			want:   true,
		},
		{
			name:   "very long maxAge - should not be stale",
			maxAge: 24 * time.Hour,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add a small delay to ensure file has some age
			time.Sleep(10 * time.Millisecond)

			got := manager.IsStale(tt.maxAge)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsStaleNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "nonexistent.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Non-existent file should return false
	isStale := manager.IsStale(1 * time.Hour)
	assert.False(t, isStale)
}

func TestForceRemove(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Create lockfile
	err := manager.Create()
	require.NoError(t, err)

	// Force remove
	err = manager.ForceRemove()
	require.NoError(t, err)

	// Verify it's gone
	assert.False(t, manager.Exists())
}

func TestForceRemoveNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "nonexistent.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Force remove non-existent file - should not error
	err := manager.ForceRemove()
	assert.NoError(t, err)
}

func TestGetPath(t *testing.T) {
	lockfilePath := "/custom/path/backup.lock"
	manager := lockfile.NewManagerWithPath(lockfilePath)

	assert.Equal(t, lockfilePath, manager.GetPath())
}

func TestLockfileLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Step 1: Verify doesn't exist
	assert.False(t, manager.Exists())

	// Step 2: Create
	err := manager.Create()
	require.NoError(t, err)
	assert.True(t, manager.Exists())

	// Step 3: Check age
	time.Sleep(50 * time.Millisecond)
	age, err := manager.GetAge()
	require.NoError(t, err)
	assert.True(t, age >= 50*time.Millisecond)

	// Step 4: Check staleness
	assert.True(t, manager.IsStale(1*time.Nanosecond), "Should be stale with tiny maxAge")
	assert.False(t, manager.IsStale(1*time.Hour), "Should not be stale with large maxAge")

	// Step 5: Remove
	err = manager.Remove()
	require.NoError(t, err)
	assert.False(t, manager.Exists())
}

func TestConcurrentLockfile(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager1 := lockfile.NewManagerWithPath(lockfilePath)
	manager2 := lockfile.NewManagerWithPath(lockfilePath)

	// First manager creates lock
	err := manager1.Create()
	require.NoError(t, err)
	defer manager1.Remove()

	// Second manager should fail to create
	err = manager2.Create()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Both managers should see it exists
	assert.True(t, manager1.Exists())
	assert.True(t, manager2.Exists())
}

func TestStaleLockfileDetection(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, "test.lock")

	manager := lockfile.NewManagerWithPath(lockfilePath)

	// Create an old lockfile by creating and modifying its timestamp
	err := manager.Create()
	require.NoError(t, err)
	defer os.Remove(lockfilePath)

	// Change modification time to 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour)
	err = os.Chtimes(lockfilePath, oldTime, oldTime)
	require.NoError(t, err)

	// Check if stale with 1 hour threshold
	isStale := manager.IsStale(1 * time.Hour)
	assert.True(t, isStale, "Lockfile should be stale (2 hours old vs 1 hour threshold)")

	// Check age
	age, err := manager.GetAge()
	require.NoError(t, err)
	assert.True(t, age >= 2*time.Hour, "Age should be at least 2 hours")
}
