## 3.1.0 (TBD)
### IMPROVEMENTS:
* **ansible-navigator Version 2 Configuration Format**: Updated YAML configuration generation to include Version 2 schema markers
  - Eliminates version migration prompts when using ansible-navigator 25.x+
  - Generated configuration files now include `ansible-navigator-settings-version: "2.0"` marker
  - Ensures `pull_policy` settings are immediately recognized without triggering migrations
  - No breaking changes - this is an internal YAML generation update transparent to users
  - Affects both `ansible-navigator` and `ansible-navigator-local` provisioners

## 3.0.0 (December 2, 2025)
### BREAKING CHANGES:
* **Provisioner Naming Swapped to Align with Packer Conventions**: The provisioner naming has been corrected to match the standard Packer pattern where SSH-based (remote) execution is the default:
  - **`ansible-navigator`** (default) now runs FROM the local machine, connects TO the target via SSH (was previously on-target execution)
  - **`ansible-navigator-local`** now runs ON the target machine with local connection (was previously the default named `ansible-navigator`)
  - This aligns with the official HashiCorp `ansible` plugin where:
    - `ansible` = SSH-based execution (default)
    - `ansible-local` = on-target execution
  - **Migration Required**: Update your Packer HCL configurations:
    ```hcl
    # Before (v2.x)
    provisioner "ansible-navigator" {
      # This ran on the target machine
    }
    provisioner "ansible-navigator-remote" {
      # This connected via SSH
    }
    
    # After (v3.x)
    provisioner "ansible-navigator-local" {
      # This runs on the target machine
    }
    provisioner "ansible-navigator" {
      # This connects via SSH (now the default)
    }
    ```
  - **Rationale**: The previous naming was inverted from Packer ecosystem conventions, causing confusion for users familiar with the official ansible plugin

### INTERNAL CHANGES:
* Directory structure updated to match new naming:
  - `provisioner/ansible-navigator-remote/` → `provisioner/ansible-navigator/`
  - `provisioner/ansible-navigator/` → `provisioner/ansible-navigator-local/`
* Go package names updated accordingly:
  - Primary provisioner now uses `package ansiblenavigator`
  - Local provisioner uses `package ansiblenavigatorlocal`
* Plugin registration updated:
  - `plugin.DEFAULT_NAME` → SSH-based provisioner (HCL: `ansible-navigator`)
  - `"local"` → on-target provisioner (HCL: `ansible-navigator-local`)

## 2.0.0 (December 2, 2025)
### BREAKING CHANGES:
* **Renamed `plays` to `play` Block Syntax**: The configuration for multiple plays has been changed from array syntax to HCL2 block syntax following Packer/Terraform conventions:
  - **Old (INCORRECT - was never valid)**: `plays = [{ ... }]`
  - **New (CORRECT)**: `play { ... }` (repeated blocks for multiple plays)
  - This aligns with HCL idioms where repeatable blocks use singular names (e.g., `provisioner`, `source`, `build`)
  - **Migration Required**: Update all configurations using play-based execution:
    ```hcl
    # Before (array syntax - this was documented but incorrect)
    plays = [
      { name = "Play 1", target = "role.one" },
      { name = "Play 2", target = "role.two" }
    ]
    
    # After (correct HCL2 block syntax)
    play {
      name   = "Play 1"
      target = "role.one"
    }
    play {
      name   = "Play 2"
      target = "role.two"
    }
    ```
  - Error messages updated to reference `play` blocks instead of `plays`
  - Internal Go field name `Plays` remains plural (it's a slice)

## 1.5.0 (December 1, 2025)
### BUG FIXES:
* **Fixed Provisioner Registration Names**: Corrected internal registration names to follow Packer SDK naming conventions:
  - Primary provisioner now uses `plugin.DEFAULT_NAME` constant (was incorrectly registering full name)
  - Remote provisioner now registers as `"remote"` (was incorrectly registering full name)
  - This fixes the awkward duplication issue where HCL references could appear as `ansible-navigator-ansible-navigator`
  - User-facing HCL syntax is unchanged: `ansible-navigator` and `ansible-navigator-remote`
* The `describe` command now correctly outputs `["-packer-default-plugin-name-", "remote"]` for provisioners

## 1.1.0 (November 3, 2025)
### NEW FEATURES:
* **JSON Output Parsing and Structured Logging**: Added support for parsing structured JSON output from ansible-navigator when running in JSON mode. This enables:
  - Fine-grained task-level failure detection
  - Improved build output clarity with per-task status reporting
  - Optional structured artifact creation for CI/CD integration
  - Real-time event streaming with detailed error summaries
* New configuration options:
  - `structured_logging`: Enable JSON event parsing and enhanced reporting (default: false)
  - `log_output_path`: Optional path to write structured summary JSON file

### IMPROVEMENTS:
* Enhanced error reporting with specific task and host failure information
* Better integration with downstream tools via JSON summary artifacts
* Improved debugging capabilities through structured event logs

## 1.0.0 (November 3, 2025)
### MAJOR CHANGES:
This release represents a complete evolution from the HashiCorp packer-plugin-ansible to the independent packer-plugin-ansible-navigator project.

#### 🔄 RENAMING AND REBRANDING:
* Replaced all identifiers, module paths, and documentation references from:
  - `packer-plugin-ansible` → `packer-plugin-ansible-navigator`
* Updated plugin registration:
  - `ansible` → `ansible-navigator`
  - `ansible-local` → `ansible-navigator-local`
* Updated Go module path to: `github.com/solomonhd/packer-plugin-ansible-navigator`

#### 📜 LICENSE CHANGE:
* Replaced Mozilla Public License 2.0 with **Apache License 2.0**
* Added Apache 2.0 headers to all Go files
* Updated copyright to SolomonHD

#### 🚀 NEW CORE FUNCTIONALITY:
* **Dual Invocation Mode**: Support for both traditional playbook files and collection plays:
  ```hcl
  # Option A: Traditional playbook file
  playbook_file = "site.yml"
  
  # Option B: Collection plays with per-play configuration (NEW)
  plays = [
    {
      name = "Migrate Node"
      target = "integration.portainer.migrate_node"
      extra_vars = {
        environment = "production"
      }
    },
    {
      name = "Configure Firewall Rules"
      target = "acme.firewall.configure_rules"
      vars_files = ["firewall.yml"]
    }
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

---

## Legacy History (HashiCorp packer-plugin-ansible)

> **Note**: The following entries are from the original [HashiCorp packer-plugin-ansible](https://github.com/hashicorp/packer-plugin-ansible) project, from which this plugin was forked. They are preserved here for historical reference only and do not apply to packer-plugin-ansible-navigator releases.

### 1.1.4 (July 30, 2025)
**IMPROVEMENTS:**

* core: added environment vars to ansible-galaxy execution
  [GH-210](https://github.com/hashicorp/packer-plugin-ansible/pull/210)
  
* docs: update links to ansible wrapper guide for clarity and fix broken links
  [GH-212](https://github.com/hashicorp/packer-plugin-ansible/pull/212)

* docs: Update ansible script link to configure remoting for ansible
  [GH-205](https://github.com/hashicorp/packer-plugin-ansible/pull/205)

* Updated plugin release process: Plugin binaries are now published on the HashiCorp official [release site](https://releases.hashicorp.com/packer-plugin-ansible), ensuring a secure and standardized delivery pipeline.

**BUG FIXES:**

* handle missing or invalid Host IP gracefully
  [GH-213](https://github.com/hashicorp/packer-plugin-ansible/pull/213)


### 1.0.0 (June 14, 2021)
The code base for this plugin has been stable since the Packer core split.
We are marking this plugin as v1.0.0 to indicate that it is stable and ready for consumption via `packer init`.

* Update packer-plugin-sdk to v0.2.3 [GH-48]
* Add module retraction for v0.0.1 as it was a bad release. [GH-46]


### 0.0.3 (May 11, 2021)

**BUG FIXES:**
* Fix registration bug that externally caused plugin not to load properly [GH-44]

### 0.0.2 (April 15, 2021)

**BUG FIXES:**
* core: Update module name in go.mod to fix plugin import path issue

### 0.0.1 (April 14, 2021)

* Ansible Plugin break out from Packer core. Changes prior to break out can be found in [Packer's CHANGELOG](https://github.com/hashicorp/packer/blob/master/CHANGELOG.md)
