// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	ansible "github.com/SolomonHD/packer-plugin-ansible-navigator/provisioner/ansible-navigator"
	ansibleLocal "github.com/SolomonHD/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local"
	"github.com/SolomonHD/packer-plugin-ansible-navigator/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterProvisioner("ansible-navigator", new(ansible.Provisioner))
	pps.RegisterProvisioner("ansible-navigator-local", new(ansibleLocal.Provisioner))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
