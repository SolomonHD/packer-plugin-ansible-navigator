// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package version

import (
	_ "embed"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/version"
)

//go:embed VERSION
var versionFile string

var (
	Version           = strings.TrimSpace(versionFile)
	VersionPrerelease = ""
	VersionMetadata   = ""
	PluginVersion     = version.NewPluginVersion(Version, VersionPrerelease, VersionMetadata)
)
