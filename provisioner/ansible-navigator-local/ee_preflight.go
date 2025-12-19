package ansiblenavigatorlocal

import (
	"fmt"
	"strings"
)

func isExecutionEnvironmentEnabled(nc *NavigatorConfig) bool {
	return nc != nil && nc.ExecutionEnvironment != nil && nc.ExecutionEnvironment.Enabled
}

// buildEEDockerPreflightShell returns a POSIX-shell snippet that prints
// debug-only EE/docker diagnostics and always succeeds.
//
// pathPrefixAssignment should be the raw PATH assignment returned by
// buildPathPrefixForRemoteShell (e.g. PATH="/opt/bin:$PATH"), or empty.
func buildEEDockerPreflightShell(pathPrefixAssignment string) string {
	prefix := ""
	if strings.TrimSpace(pathPrefixAssignment) != "" {
		prefix = pathPrefixAssignment + " "
	}

	parts := []string{
		// DOCKER_HOST (redacted)
		`if [ -z "${DOCKER_HOST+x}" ] || [ -z "$DOCKER_HOST" ]; then echo "[DEBUG] [EE preflight] DOCKER_HOST: unset"; else echo "[DEBUG] [EE preflight] DOCKER_HOST: set (redacted)"; fi`,
		// docker.sock visibility
		`if [ -S /var/run/docker.sock ]; then echo "[DEBUG] [EE preflight] /var/run/docker.sock: present (socket)"; elif [ -e /var/run/docker.sock ]; then echo "[DEBUG] [EE preflight] /var/run/docker.sock: present (not a socket)"; else echo "[DEBUG] [EE preflight] /var/run/docker.sock: missing"; fi`,
		// docker client in PATH (using the same PATH prefix semantics as ansible-navigator execution)
		fmt.Sprintf(`if %scommand -v docker >/dev/null 2>&1; then echo "[DEBUG] [EE preflight] docker client: found in PATH"; else echo "[DEBUG] [EE preflight] docker client: missing in PATH"; fi`, prefix),
		// warning-only “likely DinD” heuristic when dockerd process is detected
		fmt.Sprintf(`if %scommand -v ps >/dev/null 2>&1 && %scommand -v grep >/dev/null 2>&1; then if ps -eo comm 2>/dev/null | grep -qx dockerd; then echo "[DEBUG][WARN] [EE preflight] detected dockerd process (possible DinD); ensure DOCKER_HOST or /var/run/docker.sock is correctly wired"; fi; fi`, prefix, prefix),
	}

	return strings.Join(parts, "; ")
}
