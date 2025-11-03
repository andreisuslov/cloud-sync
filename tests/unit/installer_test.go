package unit

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/installer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExecutor is a mock implementation of CommandExecutor
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) LookPath(file string) (string, error) {
	args := m.Called(file)
	return args.String(0), args.Error(1)
}

func (m *MockExecutor) Command(name string, arg ...string) *exec.Cmd {
	args := m.Called(name, arg)
	return args.Get(0).(*exec.Cmd)
}

func (m *MockExecutor) RunCommand(cmd *exec.Cmd) error {
	args := m.Called(cmd)
	return args.Error(0)
}

func TestCheckHomebrewInstalled(t *testing.T) {
	tests := []struct {
		name     string
		lookPath string
		err      error
		want     bool
	}{
		{
			name:     "homebrew installed",
			lookPath: "/opt/homebrew/bin/brew",
			err:      nil,
			want:     true,
		},
		{
			name:     "homebrew not installed",
			lookPath: "",
			err:      errors.New("not found"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := new(MockExecutor)
			mockExec.On("LookPath", "brew").Return(tt.lookPath, tt.err)

			inst := installer.NewInstallerWithExecutor(mockExec)
			got := inst.CheckHomebrewInstalled()

			assert.Equal(t, tt.want, got)
			mockExec.AssertExpectations(t)
		})
	}
}

func TestCheckRcloneInstalled(t *testing.T) {
	tests := []struct {
		name     string
		lookPath string
		err      error
		want     bool
	}{
		{
			name:     "rclone installed",
			lookPath: "/opt/homebrew/bin/rclone",
			err:      nil,
			want:     true,
		},
		{
			name:     "rclone not installed",
			lookPath: "",
			err:      errors.New("not found"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := new(MockExecutor)
			mockExec.On("LookPath", "rclone").Return(tt.lookPath, tt.err)

			inst := installer.NewInstallerWithExecutor(mockExec)
			got := inst.CheckRcloneInstalled()

			assert.Equal(t, tt.want, got)
			mockExec.AssertExpectations(t)
		})
	}
}

func TestGetRclonePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		err     error
		wantErr bool
	}{
		{
			name:    "rclone found",
			path:    "/opt/homebrew/bin/rclone",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "rclone not found",
			path:    "",
			err:     errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := new(MockExecutor)
			mockExec.On("LookPath", "rclone").Return(tt.path, tt.err)

			inst := installer.NewInstallerWithExecutor(mockExec)
			got, err := inst.GetRclonePath()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Path may be resolved, so just check it contains rclone
				assert.Contains(t, got, "rclone")
			}
			mockExec.AssertExpectations(t)
		})
	}
}

func TestVerifyInstallation(t *testing.T) {
	tests := []struct {
		name         string
		brewFound    bool
		rcloneFound  bool
		wantErr      bool
	}{
		{
			name:         "all installed",
			brewFound:    true,
			rcloneFound:  true,
			wantErr:      false,
		},
		{
			name:         "brew missing",
			brewFound:    false,
			rcloneFound:  true,
			wantErr:      true,
		},
		{
			name:         "rclone missing",
			brewFound:    true,
			rcloneFound:  false,
			wantErr:      true,
		},
		{
			name:         "all missing",
			brewFound:    false,
			rcloneFound:  false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := new(MockExecutor)

			if tt.brewFound {
				mockExec.On("LookPath", "brew").Return("/usr/local/bin/brew", nil)
			} else {
				mockExec.On("LookPath", "brew").Return("", errors.New("not found"))
			}

			if tt.rcloneFound {
				mockExec.On("LookPath", "rclone").Return("/usr/local/bin/rclone", nil)
			} else {
				mockExec.On("LookPath", "rclone").Return("", errors.New("not found"))
			}

			inst := installer.NewInstallerWithExecutor(mockExec)
			err := inst.VerifyInstallation()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockExec.AssertExpectations(t)
		})
	}
}

func TestGetHomebrewPrefix(t *testing.T) {
	// This test just verifies the function returns a valid prefix
	prefix := installer.GetHomebrewPrefix()
	assert.NotEmpty(t, prefix)
	assert.Contains(t, []string{"/opt/homebrew", "/usr/local"}, prefix)
}

func TestGetArchitecture(t *testing.T) {
	// This test just verifies the function returns a valid architecture
	arch := installer.GetArchitecture()
	assert.NotEmpty(t, arch)
	assert.Contains(t, []string{"arm64", "amd64", "386"}, arch)
}