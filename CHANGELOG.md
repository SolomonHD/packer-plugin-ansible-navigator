## 1.0.0 (November 3, 2025)
### MAJOR CHANGES:
This release represents a complete evolution from the HashiCorp packer-plugin-ansible to the independent packer-plugin-ansible-navigator project.

#### 🔄 RENAMING AND REBRANDING:
* Replaced all identifiers, module paths, and documentation references from:
  - `packer-plugin-ansible` → `packer-plugin-ansible-navigator`
* Updated plugin registration:
  - `ansible` → `ansible-navigator`
  - `ansible-local` → `ansible-navigator-local`
* Updated Go module path to: `github.com/SolomonHD/packer-plugin-ansible-navigator`

#### 📜 LICENSE CHANGE:
* Replaced Mozilla Public License 2.0 with **Apache License 2.0**
* Added Apache 2.0 headers to all Go files
* Updated copyright to SolomonHD

#### 🚀 NEW CORE FUNCTIONALITY:
* **Dual Invocation Mode**: Support for both traditional playbook files and collection plays:
  ```hcl
  # Option A: Traditional playbook file
  playbook_file = "site.yml"
  
  # Option B: Collection plays (NEW)
  plays = [
    "integration.portainer.migrate_node",
    "acme.firewall.configure_rules"
  ]
  ```
* **Ansible Navigator Integration**: Changed from `ansible-playbook` to `ansible-navigator run`
* **Enhanced Error Handling**: 
  - Per-play failure reporting for multiple plays
  - Clear dependency checking for `ansible-navigator` binary
  - Detailed configuration validation with mutual exclusivity checks
* **Improved UI Integration**: All output properly surfaced through console and programmatic interfaces

#### ⚡ TECHNICAL IMPROVEMENTS:
* Maintained full SSH and WinRM communicator compatibility
* Enhanced error messages for better developer experience
* Updated default command to `ansible-navigator run`
* Improved validation with clear error messages:
  - "You may specify only one of `playbook_file` or `plays`"
  - "Either `playbook_file` or `plays` must be defined"
  - "ansible-navigator not found in PATH. Please install it before running this provisioner"

#### 📚 DOCUMENTATION:
* Complete rewrite of README.md with new features and examples
* Added comprehensive error handling examples
* Updated installation instructions for the new repository
* Added configuration examples for both invocation modes

## 1.1.4 (July 30, 2025)
### IMPROVEMENTS:

* core: added environment vars to ansible-galaxy execution
  [GH-210](https://github.com/hashicorp/packer-plugin-ansible/pull/210)
  
* docs: update links to ansible wrapper guide for clarity and fix broken links
  [GH-212](https://github.com/hashicorp/packer-plugin-ansible/pull/212)

* docs: Update ansible script link to configure remoting for ansible
  [GH-205](https://github.com/hashicorp/packer-plugin-ansible/pull/205)

* Updated plugin release process: Plugin binaries are now published on the HashiCorp official [release site](https://releases.hashicorp.com/packer-plugin-ansible), ensuring a secure and standardized delivery pipeline.


### BUG FIXES:

* handle missing or invalid Host IP gracefully
  [GH-213](https://github.com/hashicorp/packer-plugin-ansible/pull/213)


## 1.0.0 (June 14, 2021)
The code base for this plugin has been stable since the Packer core split.
We are marking this plugin as v1.0.0 to indicate that it is stable and ready for consumption via `packer init`.

* Update packer-plugin-sdk to v0.2.3 [GH-48]
* Add module retraction for v0.0.1 as it was a bad release. [GH-46]


## 0.0.3 (May 11, 2021)

### BUG FIXES:
* Fix registration bug that externally caused plugin not to load properly [GH-44]

## 0.0.2 (April 15, 2021)

### BUG FIXES:
* core: Update module name in go.mod to fix plugin import path issue

## 0.0.1 (April 14, 2021)

* Ansible Plugin break out from Packer core. Changes prior to break out can be found in [Packer's CHANGELOG](https://github.com/hashicorp/packer/blob/master/CHANGELOG.md)
