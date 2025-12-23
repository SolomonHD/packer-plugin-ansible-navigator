// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigatorlocal

import (
	"strings"
	"testing"
)

func TestGenerateNavigatorConfigYAML_AutomaticEEHomeXDGDefaults_WhenNotSetOrPassed(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	for _, expected := range []string{
		"HOME: /tmp",
		"XDG_CACHE_HOME: /tmp/.cache",
		"XDG_CONFIG_HOME: /tmp/.config",
	} {
		if !strings.Contains(yamlStr, expected) {
			t.Fatalf("expected %q in YAML, got: %s", expected, yamlStr)
		}
	}
}

func TestGenerateNavigatorConfigYAML_DoesNotSetHomeXDGDefaults_WhenPassedThrough(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
			EnvironmentVariables: &EnvironmentVariablesConfig{
				Pass: []string{"HOME", "XDG_CACHE_HOME", "XDG_CONFIG_HOME"},
			},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	for _, forbidden := range []string{
		"HOME: /tmp",
		"XDG_CACHE_HOME: /tmp/.cache",
		"XDG_CONFIG_HOME: /tmp/.config",
	} {
		if strings.Contains(yamlStr, forbidden) {
			t.Fatalf("did not expect %q in YAML when passed-through, got: %s", forbidden, yamlStr)
		}
	}
}

func TestGenerateNavigatorConfigYAML_DoesNotOverrideHomeXDG_WhenUserSetsValues(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
			EnvironmentVariables: &EnvironmentVariablesConfig{
				Set: map[string]string{
					"HOME":            "/custom/home",
					"XDG_CACHE_HOME":  "/custom/cache",
					"XDG_CONFIG_HOME": "/custom/config",
				},
			},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	for _, expected := range []string{
		"HOME: /custom/home",
		"XDG_CACHE_HOME: /custom/cache",
		"XDG_CONFIG_HOME: /custom/config",
	} {
		if !strings.Contains(yamlStr, expected) {
			t.Fatalf("expected %q in YAML, got: %s", expected, yamlStr)
		}
	}
}

func TestGenerateNavigatorConfigYAML_AnsibleConfigPathSchemaCompliant(t *testing.T) {
	config := &NavigatorConfig{
		Mode: "stdout",
		AnsibleConfig: &AnsibleConfig{
			Config: "/tmp/ansible.cfg",
			Defaults: &AnsibleConfigDefaults{
				RemoteTmp:       "/tmp/.ansible/tmp",
				LocalTmp:        "/tmp/.ansible-local",
				HostKeyChecking: false,
			},
			SSHConnection: &AnsibleConfigConnection{Pipelining: true},
			PrivilegeEscalation: &AnsibleConfigPrivilegeEscalation{
				Become:       true,
				BecomeMethod: "sudo",
				BecomeUser:   "root",
			},
			PersistentConnection: &AnsibleConfigPersistentConnection{ConnectTimeout: 30},
			Inventory:            &AnsibleConfigInventory{EnablePlugins: []string{"ini", "yaml"}},
			Colors:               &AnsibleConfigColors{ForceColor: true},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	if !strings.Contains(yamlStr, "path: /tmp/ansible.cfg") {
		t.Fatalf("expected ansible.config.path in YAML, got: %s", yamlStr)
	}

	for _, forbidden := range []string{
		"defaults",
		"ssh_connection",
		"privilege_escalation",
		"persistent_connection",
		"inventory",
		"paramiko_connection",
		"colors",
		"diff",
		"galaxy",
	} {
		if strings.Contains(yamlStr, forbidden) {
			t.Fatalf("did not expect %s under ansible.config in YAML, got: %s", forbidden, yamlStr)
		}
	}
}

func TestGenerateAnsibleCfgContent_NewSections(t *testing.T) {
	content, err := generateAnsibleCfgContent(&AnsibleConfig{
		PrivilegeEscalation: &AnsibleConfigPrivilegeEscalation{
			Become:       true,
			BecomeMethod: "sudo",
			BecomeUser:   "root",
		},
		PersistentConnection: &AnsibleConfigPersistentConnection{
			ConnectTimeout:      30,
			ConnectRetryTimeout: 15,
			CommandTimeout:      60,
		},
		Inventory: &AnsibleConfigInventory{EnablePlugins: []string{"ini", "yaml"}},
		ParamikoConnection: &AnsibleConfigParamikoConnection{
			ProxyCommand: "ssh -W %h:%p jumphost",
		},
		Colors: &AnsibleConfigColors{ForceColor: true},
		Diff:   &AnsibleConfigDiff{Always: true, Context: 3},
		Galaxy: &AnsibleConfigGalaxy{ServerList: []string{"automation_hub"}, IgnoreCerts: true},
	})
	if err != nil {
		t.Fatalf("generateAnsibleCfgContent failed: %v", err)
	}
	for _, expected := range []string{
		"[privilege_escalation]",
		"become = True",
		"become_method = sudo",
		"become_user = root",
		"[persistent_connection]",
		"connect_timeout = 30",
		"connect_retry_timeout = 15",
		"command_timeout = 60",
		"[inventory]",
		"enable_plugins = ini,yaml",
		"[paramiko_connection]",
		"proxy_command = ssh -W %h:%p jumphost",
		"[colors]",
		"force_color = True",
		"[diff]",
		"always = True",
		"context = 3",
		"[galaxy]",
		"server_list = automation_hub",
		"ignore_certs = True",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected %q in generated ansible.cfg, got: %q", expected, content)
		}
	}
}

func TestGenerateAnsibleCfgContent_LocalTmpIncludedWhenSet(t *testing.T) {
	content, err := generateAnsibleCfgContent(&AnsibleConfig{
		Defaults: &AnsibleConfigDefaults{
			RemoteTmp:       "/tmp/.ansible/tmp",
			LocalTmp:        "/tmp/.ansible-local",
			HostKeyChecking: false,
		},
	})
	if err != nil {
		t.Fatalf("generateAnsibleCfgContent failed: %v", err)
	}
	if !strings.Contains(content, "local_tmp = /tmp/.ansible-local") {
		t.Fatalf("expected local_tmp in generated ansible.cfg, got: %q", content)
	}
}

func TestGenerateAnsibleCfgContent_LocalTmpOmittedWhenUnset(t *testing.T) {
	content, err := generateAnsibleCfgContent(&AnsibleConfig{
		Defaults: &AnsibleConfigDefaults{
			RemoteTmp:       "/tmp/.ansible/tmp",
			HostKeyChecking: false,
		},
	})
	if err != nil {
		t.Fatalf("generateAnsibleCfgContent failed: %v", err)
	}
	if strings.Contains(content, "local_tmp") {
		t.Fatalf("did not expect local_tmp in generated ansible.cfg when unset, got: %q", content)
	}
}

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_ContainerEngine(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:         true,
			ContainerEngine: "podman",
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}
	if !strings.Contains(yamlStr, "container-engine: podman") {
		t.Fatalf("expected container-engine in YAML, got: %s", yamlStr)
	}
}

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_ContainerOptions(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:          true,
			ContainerOptions: []string{"--net=host"},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}
	if !strings.Contains(yamlStr, "container-options:") {
		t.Fatalf("expected container-options in YAML, got: %s", yamlStr)
	}
	if !strings.Contains(yamlStr, "- --net=host") {
		t.Fatalf("expected container option value in YAML, got: %s", yamlStr)
	}
}

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_PullArgumentsOnly(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:       true,
			PullArguments: []string{"--tls-verify=false"},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}
	if !strings.Contains(yamlStr, "pull:") || !strings.Contains(yamlStr, "arguments:") {
		t.Fatalf("expected pull.arguments in YAML, got: %s", yamlStr)
	}
	if !strings.Contains(yamlStr, "- --tls-verify=false") {
		t.Fatalf("expected pull argument list item in YAML, got: %s", yamlStr)
	}
}
