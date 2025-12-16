// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigatorlocal

import (
	"fmt"
	"os"
	"strings"
)

func needsGeneratedAnsibleCfg(ansibleCfg *AnsibleConfig) bool {
	if ansibleCfg == nil {
		return false
	}
	return ansibleCfg.Defaults != nil || ansibleCfg.SSHConnection != nil
}

func formatAnsibleCfgBool(v bool) string {
	if v {
		return "True"
	}
	return "False"
}

func generateAnsibleCfgContent(ansibleCfg *AnsibleConfig) (string, error) {
	if ansibleCfg == nil {
		return "", nil
	}

	var b strings.Builder
	wroteAny := false

	if ansibleCfg.Defaults != nil {
		b.WriteString("[defaults]\n")
		wroteAny = true

		if ansibleCfg.Defaults.RemoteTmp != "" {
			b.WriteString(fmt.Sprintf("remote_tmp = %s\n", ansibleCfg.Defaults.RemoteTmp))
		}
		if ansibleCfg.Defaults.LocalTmp != "" {
			b.WriteString(fmt.Sprintf("local_tmp = %s\n", ansibleCfg.Defaults.LocalTmp))
		}
		b.WriteString(fmt.Sprintf("host_key_checking = %s\n", formatAnsibleCfgBool(ansibleCfg.Defaults.HostKeyChecking)))
		b.WriteString("\n")
	}

	if ansibleCfg.SSHConnection != nil {
		b.WriteString("[ssh_connection]\n")
		wroteAny = true

		if ansibleCfg.SSHConnection.SSHTimeout > 0 {
			b.WriteString(fmt.Sprintf("ssh_timeout = %d\n", ansibleCfg.SSHConnection.SSHTimeout))
		}
		b.WriteString(fmt.Sprintf("pipelining = %s\n", formatAnsibleCfgBool(ansibleCfg.SSHConnection.Pipelining)))
		b.WriteString("\n")
	}

	if !wroteAny {
		return "", nil
	}

	return b.String(), nil
}

func createTempAnsibleCfgFile(content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("ansible.cfg content cannot be empty")
	}

	tmpFile, err := os.CreateTemp("", "packer-ansible-cfg-*.ini")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary ansible.cfg file: %w", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write ansible.cfg content: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close ansible.cfg temp file: %w", err)
	}

	return tmpFile.Name(), nil
}
