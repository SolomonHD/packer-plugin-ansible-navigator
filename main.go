// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	ansiblenavigator "github.com/solomonhd/packer-plugin-ansible-navigator/provisioner/ansible-navigator"
	ansiblenavigatorlocal "github.com/solomonhd/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local"
	"github.com/solomonhd/packer-plugin-ansible-navigator/version"
)

func main() {
	pps := plugin.NewSet()
	// Register provisioners using Packer SDK naming conventions:
	// - plugin.DEFAULT_NAME ("-packer-default-plugin-name-") for primary provisioner (SSH-based)
	//   -> accessible in HCL as "ansible-navigator"
	// - "local" for secondary provisioner (runs on target)
	//   -> accessible in HCL as "ansible-navigator-local" (Packer prefixes with plugin alias)
	pps.RegisterProvisioner(plugin.DEFAULT_NAME, new(ansiblenavigator.Provisioner))
	pps.RegisterProvisioner("local", new(ansiblenavigatorlocal.Provisioner))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
