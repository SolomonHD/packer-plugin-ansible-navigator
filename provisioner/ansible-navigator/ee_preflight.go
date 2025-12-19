package ansiblenavigator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func debugWarnf(ui packersdk.Ui, enabled bool, format string, args ...interface{}) {
	if !enabled {
		return
	}
	ui.Message(fmt.Sprintf("[DEBUG][WARN] "+format, args...))
}

func isExecutionEnvironmentEnabled(nc *NavigatorConfig) bool {
	return nc != nil && nc.ExecutionEnvironment != nil && nc.ExecutionEnvironment.Enabled
}

type eeDockerPreflightDeps struct {
	stat               func(string) (os.FileInfo, error)
	lookPath           func(string) (string, error)
	commandOutput      func(ctx context.Context, name string, args ...string) ([]byte, error)
	dockerHostProvider func() (string, bool) // value, present
}

func defaultEEDockerPreflightDeps() eeDockerPreflightDeps {
	return eeDockerPreflightDeps{
		stat:     os.Stat,
		lookPath: exec.LookPath,
		commandOutput: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			return cmd.Output()
		},
		dockerHostProvider: func() (string, bool) {
			v, ok := os.LookupEnv("DOCKER_HOST")
			return v, ok
		},
	}
}

type eeDockerPreflightResult struct {
	dockerHostPresent bool
	dockerHostValue   string

	dockerSockExists      bool
	dockerSockIsSocket    bool
	dockerSockStatErrored bool

	dockerClientFound bool

	dockerdDetected      bool
	dockerdCheckPossible bool
}

func detectDockerdFast(deps eeDockerPreflightDeps) (detected bool, possible bool) {
	// Guard: if ps isn't present, skip silently.
	if _, err := deps.lookPath("ps"); err != nil {
		return false, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	out, err := deps.commandOutput(ctx, "ps", "-eo", "comm")
	if err != nil {
		return false, true
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "dockerd" {
			return true, true
		}
	}
	return false, true
}

func findExecutableInPath(stat func(string) (os.FileInfo, error), pathValue, exe string) bool {
	// Use a shell-style PATH search without mutating process env.
	for _, dir := range strings.Split(pathValue, string(os.PathListSeparator)) {
		if dir == "" {
			continue
		}
		candidate := dir + string(os.PathSeparator) + exe
		info, err := stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		// Any execute bit on Unix.
		if info.Mode()&0111 != 0 {
			return true
		}
	}
	return false
}

func collectEEDockerPreflight(ansibleNavigatorPath []string, deps eeDockerPreflightDeps) eeDockerPreflightResult {
	var res eeDockerPreflightResult

	res.dockerHostValue, res.dockerHostPresent = deps.dockerHostProvider()

	info, err := deps.stat("/var/run/docker.sock")
	if err == nil {
		res.dockerSockExists = true
		res.dockerSockIsSocket = info.Mode()&os.ModeSocket != 0
	} else if !os.IsNotExist(err) {
		res.dockerSockStatErrored = true
	}

	lookupPath := os.Getenv("PATH")
	if len(ansibleNavigatorPath) > 0 {
		prefix := strings.Join(ansibleNavigatorPath, string(os.PathListSeparator))
		if prefix != "" {
			lookupPath = prefix + string(os.PathListSeparator) + lookupPath
		}
	}
	res.dockerClientFound = findExecutableInPath(deps.stat, lookupPath, "docker")

	res.dockerdDetected, res.dockerdCheckPossible = detectDockerdFast(deps)

	return res
}

func emitEEDockerPreflight(ui packersdk.Ui, debugEnabled bool, ansibleNavigatorPath []string) {
	if !debugEnabled {
		return
	}

	deps := defaultEEDockerPreflightDeps()
	res := collectEEDockerPreflight(ansibleNavigatorPath, deps)

	// DOCKER_HOST
	if !res.dockerHostPresent || strings.TrimSpace(res.dockerHostValue) == "" {
		debugf(ui, true, "[EE preflight] DOCKER_HOST: unset")
	} else {
		// Avoid leaking secrets (user:pass@, tokens, etc.).
		debugf(ui, true, "[EE preflight] DOCKER_HOST: set (redacted)")
	}

	// docker.sock
	if res.dockerSockStatErrored {
		debugf(ui, true, "[EE preflight] /var/run/docker.sock: stat error")
	} else if !res.dockerSockExists {
		debugf(ui, true, "[EE preflight] /var/run/docker.sock: missing")
	} else if res.dockerSockIsSocket {
		debugf(ui, true, "[EE preflight] /var/run/docker.sock: present (socket)")
	} else {
		debugf(ui, true, "[EE preflight] /var/run/docker.sock: present (not a socket)")
	}

	// docker client
	if res.dockerClientFound {
		debugf(ui, true, "[EE preflight] docker client: found in PATH")
	} else {
		debugf(ui, true, "[EE preflight] docker client: missing in PATH")
	}

	// dockerd heuristic (warning-only)
	if res.dockerdDetected {
		debugWarnf(ui, true, "[EE preflight] detected dockerd process (possible DinD); ensure DOCKER_HOST or /var/run/docker.sock is correctly wired")
	}
}

// emitEEDockerPreflightWithDeps is a test seam.
func emitEEDockerPreflightWithDeps(ui packersdk.Ui, debugEnabled bool, ansibleNavigatorPath []string, deps eeDockerPreflightDeps) {
	if !debugEnabled {
		return
	}
	res := collectEEDockerPreflight(ansibleNavigatorPath, deps)

	if !res.dockerHostPresent || strings.TrimSpace(res.dockerHostValue) == "" {
		debugf(ui, true, "[EE preflight] DOCKER_HOST: unset")
	} else {
		debugf(ui, true, "[EE preflight] DOCKER_HOST: set (redacted)")
	}

	if res.dockerSockStatErrored {
		debugf(ui, true, "[EE preflight] /var/run/docker.sock: stat error")
	} else if !res.dockerSockExists {
		debugf(ui, true, "[EE preflight] /var/run/docker.sock: missing")
	} else if res.dockerSockIsSocket {
		debugf(ui, true, "[EE preflight] /var/run/docker.sock: present (socket)")
	} else {
		debugf(ui, true, "[EE preflight] /var/run/docker.sock: present (not a socket)")
	}

	if res.dockerClientFound {
		debugf(ui, true, "[EE preflight] docker client: found in PATH")
	} else {
		debugf(ui, true, "[EE preflight] docker client: missing in PATH")
	}

	if res.dockerdDetected {
		debugWarnf(ui, true, "[EE preflight] detected dockerd process (possible DinD); ensure DOCKER_HOST or /var/run/docker.sock is correctly wired")
	}
}
