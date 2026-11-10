// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandUserPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err, "Failed to get home directory")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bare tilde expands to HOME",
			input:    "~",
			expected: home,
		},
		{
			name:     "Tilde with subdirectory",
			input:    "~/ansible/playbooks",
			expected: filepath.Join(home, "ansible/playbooks"),
		},
		{
			name:     "Tilde with single level",
			input:    "~/bin",
			expected: filepath.Join(home, "bin"),
		},
		{
			name:     "Multi-user home preserved",
			input:    "~otheruser/files",
			expected: "~otheruser/files",
		},
		{
			name:     "Absolute path unchanged",
			input:    "/usr/local/bin/ansible",
			expected: "/usr/local/bin/ansible",
		},
		{
			name:     "Relative path unchanged",
			input:    "./local/ansible",
			expected: "./local/ansible",
		},
		{
			name:     "Empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "Just filename unchanged",
			input:    "ansible-navigator",
			expected: "ansible-navigator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandUserPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildEnvWithPath(t *testing.T) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	tests := []struct {
		name                 string
		ansibleNavigatorPath []string
		wantPathPrefix       []string
		wantContainsOriginal bool
	}{
		{
			name:                 "Empty path list returns original environment",
			ansibleNavigatorPath: []string{},
			wantPathPrefix:       nil,
			wantContainsOriginal: true,
		},
		{
			name:                 "Single directory prepended",
			ansibleNavigatorPath: []string{"/opt/ansible/bin"},
			wantPathPrefix:       []string{"/opt/ansible/bin"},
			wantContainsOriginal: true,
		},
		{
			name:                 "Multiple directories prepended in order",
			ansibleNavigatorPath: []string{"/opt/ansible/bin", "/usr/local/ansible/bin"},
			wantPathPrefix:       []string{"/opt/ansible/bin", "/usr/local/ansible/bin"},
			wantContainsOriginal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := buildEnvWithPath(tt.ansibleNavigatorPath)

			// Find PATH in environment
			var pathValue string
			for _, envVar := range env {
				if strings.HasPrefix(envVar, "PATH=") {
					pathValue = strings.TrimPrefix(envVar, "PATH=")
					break
				}
			}

			if len(tt.wantPathPrefix) == 0 {
				// Should contain original PATH
				if tt.wantContainsOriginal {
					assert.Contains(t, pathValue, originalPath)
				}
			} else {
				// Check that new directories are at the beginning
				for _, dir := range tt.wantPathPrefix {
					assert.Contains(t, pathValue, dir)
				}

				// Check they appear before original PATH
				if tt.wantContainsOriginal && originalPath != "" {
					assert.Contains(t, pathValue, originalPath)
					// First entry should be from our list
					pathParts := strings.Split(pathValue, string(os.PathListSeparator))
					assert.GreaterOrEqual(t, len(pathParts), len(tt.wantPathPrefix))
					for i, expectedDir := range tt.wantPathPrefix {
						assert.Equal(t, expectedDir, pathParts[i])
					}
				}
			}
		})
	}
}

func TestConfigValidateCommandWithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		command string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid command without spaces",
			command: "ansible-navigator",
			wantErr: false,
		},
		{
			name:    "Valid command as path",
			command: "/usr/local/bin/ansible-navigator",
			wantErr: false,
		},
		{
			name:    "Invalid command with space",
			command: "ansible-navigator run",
			wantErr: true,
			errMsg:  "command must be only the executable name or path (no arguments)",
		},
		{
			name:    "Invalid command with multiple spaces",
			command: "ansible-navigator --mode json",
			wantErr: true,
			errMsg:  "Found whitespace in",
		},
		{
			name:    "Invalid command with tab",
			command: "ansible-navigator\trun",
			wantErr: true,
			errMsg:  "command must be only the executable name or path",
		},
		{
			name:    "Invalid command with leading space",
			command: " ansible-navigator",
			wantErr: true,
			errMsg:  "Found whitespace in",
		},
		{
			name:    "Invalid command with trailing space",
			command: "ansible-navigator ",
			wantErr: true,
			errMsg:  "Found whitespace in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Command: tt.command,
				Plays:   []Play{{Target: validPlaybook}},
			}

			err := cfg.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrepareWithAnsibleNavigatorPath(t *testing.T) {
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name                 string
		ansibleNavigatorPath []string
		wantExpandedPath     []string
	}{
		{
			name:                 "Absolute paths unchanged",
			ansibleNavigatorPath: []string{"/opt/ansible/bin", "/usr/local/bin"},
			wantExpandedPath:     []string{"/opt/ansible/bin", "/usr/local/bin"},
		},
		{
			name:                 "Tilde paths expanded",
			ansibleNavigatorPath: []string{"~/bin", "~/ansible/bin"},
			wantExpandedPath:     []string{filepath.Join(home, "bin"), filepath.Join(home, "ansible/bin")},
		},
		{
			name:                 "Mixed absolute and tilde paths",
			ansibleNavigatorPath: []string{"~/bin", "/opt/ansible/bin"},
			wantExpandedPath:     []string{filepath.Join(home, "bin"), "/opt/ansible/bin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{}
			config := map[string]interface{}{
				"play":                   []map[string]interface{}{{"target": validPlaybook}},
				"skip_version_check":     true,
				"ansible_navigator_path": tt.ansibleNavigatorPath,
			}

			err := p.Prepare(config)
			require.NoError(t, err)

			assert.Equal(t, tt.wantExpandedPath, p.config.AnsibleNavigatorPath)
		})
	}
}

func TestPreparePathExpansion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	// Create a playbook in HOME so we can reference it via ~
	homePlaybook, err := os.CreateTemp(home, "packer-ansible-navigator-home-playbook-*.yml")
	require.NoError(t, err)
	_, _ = homePlaybook.WriteString("---\n- hosts: all\n")
	require.NoError(t, homePlaybook.Close())
	defer os.Remove(homePlaybook.Name())

	tildePlaybook := filepath.ToSlash(filepath.Join("~", filepath.Base(homePlaybook.Name())))
	absPlaybook := filepath.Join(home, filepath.Base(homePlaybook.Name()))

	p := &Provisioner{}
	config := map[string]interface{}{
		"skip_version_check": true,
		"play":               []map[string]interface{}{{"target": tildePlaybook}},
		"command":            "~/bin/ansible-navigator",
	}

	require.NoError(t, p.Prepare(config))
	require.Equal(t, filepath.Join(home, "bin/ansible-navigator"), p.config.Command)
	require.Len(t, p.config.Plays, 1)
	require.Equal(t, absPlaybook, p.config.Plays[0].Target)
}

// Regression: exec.Command searches the parent's PATH before cmd.Env is assigned.
func TestConfiguredPathControlsVersionAndPlayLaunch(t *testing.T) {
	requirePOSIXShellFixture(t)
	for _, site := range []string{"version", "play"} {
		t.Run(site, func(t *testing.T) {
			dir := t.TempDir()
			logPath := filepath.Join(dir, "calls")
			script := "#!/bin/sh\nprintf '%s\\n' \"$0 $1\" >> \"$CALL_LOG\"\necho 'ansible-navigator 26.11.0'\n"
			require.NoError(t, os.WriteFile(filepath.Join(dir, "ansible-navigator"), []byte(script), 0755))
			t.Setenv("PATH", t.TempDir())
			t.Setenv("CALL_LOG", logPath)
			playbook := filepath.Join(dir, "site.yml")
			require.NoError(t, os.WriteFile(playbook, []byte("- hosts: all\n"), 0600))
			p := &Provisioner{}
			cfg := map[string]interface{}{"play": []map[string]interface{}{{"target": playbook}}, "ansible_navigator_path": []string{dir}, "skip_version_check": site == "play"}
			require.NoError(t, p.Prepare(cfg))
			if site == "play" {
				var output bytes.Buffer
				ui := &packersdk.BasicUi{Writer: &output, ErrorWriter: &output}
				require.NoError(t, p.executePlays(ui, nil, "", "", "", ""))
			}
			calls, err := os.ReadFile(logPath)
			require.NoError(t, err)
			want := "--version"
			if site == "play" {
				want = "run"
			}
			require.Contains(t, string(calls), filepath.Join(dir, "ansible-navigator")+" "+want)
		})
	}
}

func requirePOSIXShellFixture(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fixture requires native POSIX shebang execution")
	}
}

func navigatorFixture(t *testing.T, dir, name, body string) string {
	t.Helper()
	requirePOSIXShellFixture(t)
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0755))
	return path
}

func TestNavigatorExecutableSelection(t *testing.T) {
	home, first, second, inherited := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", inherited)
	require.NoError(t, os.Mkdir(filepath.Join(home, "bin"), 0700))
	for _, dir := range []string{first, second, inherited, filepath.Join(home, "bin")} {
		navigatorFixture(t, dir, "ansible-navigator", "echo 'ansible-navigator 26.11.0'")
	}
	cwd, err := os.Getwd()
	require.NoError(t, err)
	rel, err := filepath.Rel(cwd, filepath.Join(inherited, "ansible-navigator"))
	require.NoError(t, err)
	for _, tc := range []struct {
		name, command string
		paths         []string
		want          string
	}{
		{"ordered configured paths", "ansible-navigator", []string{first, second}, filepath.Join(first, "ansible-navigator")},
		{"inherited fallback", "ansible-navigator", []string{t.TempDir()}, filepath.Join(inherited, "ansible-navigator")},
		{"inherited only", "ansible-navigator", nil, filepath.Join(inherited, "ansible-navigator")},
		{"absolute authoritative", filepath.Join(inherited, "ansible-navigator"), []string{first}, filepath.Join(inherited, "ansible-navigator")},
		{"relative authoritative", rel, []string{first}, filepath.Join(inherited, "ansible-navigator")},
		{"home command", "~/bin/ansible-navigator", []string{first}, filepath.Join(home, "bin/ansible-navigator")},
		{"home search path", "ansible-navigator", []string{"~/bin"}, filepath.Join(home, "bin/ansible-navigator")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &Provisioner{}
			require.NoError(t, p.Prepare(map[string]interface{}{"command": tc.command, "ansible_navigator_path": tc.paths, "play": []map[string]interface{}{{"target": "example.role"}}}))
			require.Equal(t, tc.want, p.resolvedCommand)
			require.Equal(t, inherited, os.Getenv("PATH"))
		})
	}
	nonExecutable := filepath.Join(first, "not-executable")
	require.NoError(t, os.WriteFile(nonExecutable, []byte("ignored"), 0600))
	for _, command := range []string{nonExecutable, first, "missing-navigator", "ansible-navigator --version"} {
		p := &Provisioner{config: Config{Command: command}}
		require.Error(t, p.resolveExecutable(context.Background()))
	}
}

func TestNavigatorReusesSelectedExecutableForEveryPlay(t *testing.T) {
	for _, manager := range []string{"", "asdf", "rbenv", "pyenv"} {
		t.Run(manager, func(t *testing.T) {
			dir, conflict := t.TempDir(), t.TempDir()
			calls, managerCalls := filepath.Join(dir, "calls"), filepath.Join(dir, "manager-calls")
			t.Setenv("CALL_LOG", calls)
			t.Setenv("MANAGER_LOG", managerCalls)
			t.Setenv("PATH", conflict)
			realName := "real-navigator"
			if manager == "" {
				realName = "ansible-navigator"
			}
			real := navigatorFixture(t, dir, realName, `printf '%s %s\n' "$0" "$1" >> "$CALL_LOG"
echo 'ansible-navigator 26.11.0'`)
			command := "ansible-navigator"
			if manager != "" {
				command = navigatorFixture(t, dir, "ansible-navigator", "# "+manager+" exec\nexit 99")
				navigatorFixture(t, dir, manager, fmt.Sprintf("echo called >> \"$MANAGER_LOG\"\nprintf '%%s\\n' '%s'", real))
			}
			navigatorFixture(t, conflict, "ansible-navigator", "exit 98")
			playbook := filepath.Join(dir, "site.yml")
			require.NoError(t, os.WriteFile(playbook, []byte("- hosts: all\n"), 0600))
			p := &Provisioner{}
			require.NoError(t, p.Prepare(map[string]interface{}{"command": command, "ansible_navigator_path": []string{dir}, "play": []map[string]interface{}{{"target": playbook}, {"target": playbook}}}))
			// Even a changed command installation selection cannot re-enter the shim.
			p.config.Command = filepath.Join(conflict, "ansible-navigator")
			var output bytes.Buffer
			ui := &packersdk.BasicUi{Writer: &output, ErrorWriter: &output}
			require.NoError(t, p.executePlays(ui, nil, "", "", "", ""))
			got, err := os.ReadFile(calls)
			require.NoError(t, err)
			require.Equal(t, real+" --version\n"+real+" run\n"+real+" run\n", string(got))
			if manager != "" {
				got, err = os.ReadFile(managerCalls)
				require.NoError(t, err)
				require.Equal(t, "called\n", string(got))
			}
		})
	}
}

func TestShimResolutionFailuresAndBudgets(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"manager failure", "exit 7", "subprocess failed"},
		{"empty result", "exit 0", "empty path"},
		{"relative result", "echo relative", "non-absolute"},
		{"missing result", "echo /no/such/navigator", "invalid executable"},
		{"non executable result", "echo \"$BAD_EXECUTABLE\"", "invalid executable"},
		{"unresolved shim chain", "echo \"$SHIM_EXECUTABLE\"", "unresolved shim chain"},
		{"hanging manager", "exec /bin/sleep 3", "timed out"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("PATH", t.TempDir())
			shim := navigatorFixture(t, dir, "ansible-navigator", "# asdf exec\nexit 99")
			bad := filepath.Join(dir, "not-executable")
			require.NoError(t, os.WriteFile(bad, []byte("x"), 0600))
			t.Setenv("BAD_EXECUTABLE", bad)
			t.Setenv("SHIM_EXECUTABLE", shim)
			navigatorFixture(t, dir, "asdf", tc.body)
			p := &Provisioner{config: Config{Command: "ansible-navigator", AnsibleNavigatorPath: []string{dir}, VersionCheckTimeout: stringPtr("120ms")}}
			start := time.Now()
			err := p.getVersion()
			require.ErrorContains(t, err, tc.want)
			require.ErrorContains(t, err, "real executable's full path")
			require.Less(t, time.Since(start), time.Second)
		})
	}
	t.Run("shared version budget", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("PATH", dir)
		real := navigatorFixture(t, dir, "real", "exec /bin/sleep 3")
		navigatorFixture(t, dir, "ansible-navigator", "# pyenv exec\nexit 99")
		navigatorFixture(t, dir, "pyenv", "/bin/sleep 0.3\necho '"+real+"'")
		p := &Provisioner{config: Config{Command: "ansible-navigator", VersionCheckTimeout: stringPtr("500ms")}}
		start := time.Now()
		require.ErrorContains(t, p.getVersion(), "version check timed out")
		require.Less(t, time.Since(start), 700*time.Millisecond)
	})
	t.Run("skipped version retains independent resolution", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("PATH", dir)
		real := navigatorFixture(t, dir, "real", `[ "$1" = run ] || exit 97`)
		navigatorFixture(t, dir, "ansible-navigator", "# rbenv exec\nexit 99")
		navigatorFixture(t, dir, "rbenv", "/bin/sleep 0.05\necho '"+real+"'")
		playbook := filepath.Join(dir, "site.yml")
		require.NoError(t, os.WriteFile(playbook, []byte("- hosts: all\n"), 0600))
		p := &Provisioner{}
		require.NoError(t, p.Prepare(map[string]interface{}{"skip_version_check": true, "version_check_timeout": "1ns", "play": []map[string]interface{}{{"target": playbook}}}))
		var output bytes.Buffer
		ui := &packersdk.BasicUi{Writer: &output, ErrorWriter: &output}
		require.NoError(t, p.executePlays(ui, nil, "", "", "", ""))
		require.Equal(t, real, p.resolvedCommand)
	})
	t.Run("independent resolution deadline", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("PATH", dir)
		navigatorFixture(t, dir, "ansible-navigator", "# asdf exec\nexit 99")
		navigatorFixture(t, dir, "asdf", "exec /bin/sleep 30")
		p := &Provisioner{config: Config{Command: "ansible-navigator", SkipVersionCheck: true, VersionCheckTimeout: stringPtr("1ns")}}
		start := time.Now()
		require.ErrorContains(t, p.resolveExecutable(context.Background()), "timed out")
		require.GreaterOrEqual(t, time.Since(start), 9*time.Second)
		require.Less(t, time.Since(start), 11*time.Second)
	})
}

func TestExplicitRelativeCommandUsesPreparationDirectory(t *testing.T) {
	dir := t.TempDir()
	real := navigatorFixture(t, dir, "ansible-navigator", "exit 0")
	p := &Provisioner{}
	cwd, err := os.Getwd()
	require.NoError(t, err)
	relative, err := filepath.Rel(cwd, real)
	require.NoError(t, err)
	require.NoError(t, p.Prepare(map[string]interface{}{"command": relative, "skip_version_check": true, "play": []map[string]interface{}{{"target": "example.role"}}}))
	// Resolve after a working-directory change, as skipped version checks do.
	require.NoError(t, os.Chdir(t.TempDir()))
	defer os.Chdir(cwd)
	require.NoError(t, p.resolveExecutable(context.Background()))
	require.Equal(t, real, p.resolvedCommand)
}

func TestMissingShimManagerIsActionable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	navigatorFixture(t, dir, "ansible-navigator", "# asdf exec\nexit 99")
	p := &Provisioner{config: Config{Command: "ansible-navigator"}}
	err := p.resolveExecutable(context.Background())
	require.ErrorContains(t, err, "asdf manager lookup failed")
	require.ErrorContains(t, err, "real executable's full path")
}
