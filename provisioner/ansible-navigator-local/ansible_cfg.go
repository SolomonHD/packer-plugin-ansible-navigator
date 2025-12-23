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
	return ansibleCfg.Defaults != nil ||
		ansibleCfg.SSHConnection != nil ||
		ansibleCfg.PrivilegeEscalation != nil ||
		ansibleCfg.PersistentConnection != nil ||
		ansibleCfg.Inventory != nil ||
		ansibleCfg.ParamikoConnection != nil ||
		ansibleCfg.Colors != nil ||
		ansibleCfg.Diff != nil ||
		ansibleCfg.Galaxy != nil
}

func formatAnsibleCfgBool(v bool) string {
	if v {
		return "True"
	}
	return "False"
}

func formatAnsibleCfgList(v []string) string {
	// Ansible config list values are generally represented as comma-separated strings.
	// Preserve the provided ordering for determinism.
	return strings.Join(v, ",")
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

	if ansibleCfg.PrivilegeEscalation != nil {
		b.WriteString("[privilege_escalation]\n")
		wroteAny = true

		b.WriteString(fmt.Sprintf("become = %s\n", formatAnsibleCfgBool(ansibleCfg.PrivilegeEscalation.Become)))
		if ansibleCfg.PrivilegeEscalation.BecomeMethod != "" {
			b.WriteString(fmt.Sprintf("become_method = %s\n", ansibleCfg.PrivilegeEscalation.BecomeMethod))
		}
		if ansibleCfg.PrivilegeEscalation.BecomeUser != "" {
			b.WriteString(fmt.Sprintf("become_user = %s\n", ansibleCfg.PrivilegeEscalation.BecomeUser))
		}
		b.WriteString("\n")
	}

	if ansibleCfg.PersistentConnection != nil {
		b.WriteString("[persistent_connection]\n")
		wroteAny = true

		if ansibleCfg.PersistentConnection.ConnectTimeout > 0 {
			b.WriteString(fmt.Sprintf("connect_timeout = %d\n", ansibleCfg.PersistentConnection.ConnectTimeout))
		}
		if ansibleCfg.PersistentConnection.ConnectRetryTimeout > 0 {
			b.WriteString(fmt.Sprintf("connect_retry_timeout = %d\n", ansibleCfg.PersistentConnection.ConnectRetryTimeout))
		}
		if ansibleCfg.PersistentConnection.CommandTimeout > 0 {
			b.WriteString(fmt.Sprintf("command_timeout = %d\n", ansibleCfg.PersistentConnection.CommandTimeout))
		}
		b.WriteString("\n")
	}

	if ansibleCfg.Inventory != nil {
		b.WriteString("[inventory]\n")
		wroteAny = true

		if len(ansibleCfg.Inventory.EnablePlugins) > 0 {
			b.WriteString(fmt.Sprintf("enable_plugins = %s\n", formatAnsibleCfgList(ansibleCfg.Inventory.EnablePlugins)))
		}
		b.WriteString("\n")
	}

	if ansibleCfg.ParamikoConnection != nil {
		b.WriteString("[paramiko_connection]\n")
		wroteAny = true

		if ansibleCfg.ParamikoConnection.ProxyCommand != "" {
			b.WriteString(fmt.Sprintf("proxy_command = %s\n", ansibleCfg.ParamikoConnection.ProxyCommand))
		}
		b.WriteString("\n")
	}

	if ansibleCfg.Colors != nil {
		b.WriteString("[colors]\n")
		wroteAny = true

		b.WriteString(fmt.Sprintf("force_color = %s\n", formatAnsibleCfgBool(ansibleCfg.Colors.ForceColor)))
		b.WriteString("\n")
	}

	if ansibleCfg.Diff != nil {
		b.WriteString("[diff]\n")
		wroteAny = true

		b.WriteString(fmt.Sprintf("always = %s\n", formatAnsibleCfgBool(ansibleCfg.Diff.Always)))
		if ansibleCfg.Diff.Context > 0 {
			b.WriteString(fmt.Sprintf("context = %d\n", ansibleCfg.Diff.Context))
		}
		b.WriteString("\n")
	}

	if ansibleCfg.Galaxy != nil {
		b.WriteString("[galaxy]\n")
		wroteAny = true

		if len(ansibleCfg.Galaxy.ServerList) > 0 {
			b.WriteString(fmt.Sprintf("server_list = %s\n", formatAnsibleCfgList(ansibleCfg.Galaxy.ServerList)))
		}
		b.WriteString(fmt.Sprintf("ignore_certs = %s\n", formatAnsibleCfgBool(ansibleCfg.Galaxy.IgnoreCerts)))
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
