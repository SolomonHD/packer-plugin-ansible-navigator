# 📚 Packer Plugin Ansible Navigator Documentation

Welcome to the comprehensive documentation for the Packer Plugin Ansible Navigator.

## Documentation Index

### 🚀 Getting Started

- **[Main README](../README.md)** - Project overview and quick start guide
- **[Installation Guide](INSTALLATION.md)** - All installation methods and requirements
- **[Configuration Reference](CONFIGURATION.md)** - Complete list of options and parameters

### 📖 Detailed Guides

- **[Examples Gallery](EXAMPLES.md)** - Real-world examples and use cases
- **[Troubleshooting Guide](TROUBLESHOOTING.md)** - Common issues and solutions
- **[JSON Logging](JSON_LOGGING.md)** - Structured logging for automation
- **[Collection Plays](UNIFIED_PLAYS.md)** - Using Ansible Collection plays

## Quick Reference

### Essential Links

- **Repository**: [github.com/solomonhd/packer-plugin-ansible-navigator](https://github.com/solomonhd/packer-plugin-ansible-navigator)
- **Issues**: [GitHub Issues](https://github.com/solomonhd/packer-plugin-ansible-navigator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/solomonhd/packer-plugin-ansible-navigator/discussions)
- **Releases**: [GitHub Releases](https://github.com/solomonhd/packer-plugin-ansible-navigator/releases)

### Plugin Information

- **Current Version**: 1.1.0
- **License**: Apache License 2.0
- **Go Version Required**: >= 1.25.3
- **Packer Version Required**: >= 1.7.0

## What is Ansible Navigator?

Ansible Navigator is a modern interface for running Ansible automation that provides:

- **Containerized Execution**: Run playbooks in isolated execution environments
- **Consistent Experience**: Same behavior across different systems
- **Enhanced Security**: Isolated runtime environments
- **Better Debugging**: Interactive and structured output modes

This Packer plugin integrates Ansible Navigator into your image building workflow, allowing you to:

- Use containerized Ansible execution environments
- Run Ansible Collection plays directly
- Get detailed JSON logging and error reporting
- Maintain consistency across development and production

## Documentation Structure

### For New Users

1. Start with the [Installation Guide](INSTALLATION.md)
2. Review basic examples in the [Examples Gallery](EXAMPLES.md)
3. Explore the [Configuration Reference](CONFIGURATION.md) for available options

### For Migration from ansible-playbook

If you're migrating from the traditional Ansible provisioner:

1. Review [Collection Plays](UNIFIED_PLAYS.md) for the new execution model
2. Check the [Configuration Reference](CONFIGURATION.md) for mapping old options
3. See [Troubleshooting](TROUBLESHOOTING.md) for common migration issues

### For Advanced Users

- [JSON Logging](JSON_LOGGING.md) - Set up structured logging for CI/CD
- [Collection Plays](UNIFIED_PLAYS.md) - Advanced play execution patterns
- [Examples Gallery](EXAMPLES.md#production-patterns) - Production-ready patterns

## Key Features

### 🎯 Dual Invocation Mode

Choose between traditional playbooks or modern collection plays:

```hcl
# Traditional playbook
playbook_file = "site.yml"

# OR Collection plays (array of objects)
plays = [
  {
    name = "My Play"
    target = "namespace.collection.play_name"
    extra_vars = {}  # Optional per-play variables
  }
]
```

### 📦 Execution Environments

Use containerized Ansible environments for consistency:

```hcl
execution_environment = "quay.io/ansible/creator-ee:latest"
```

### 📊 Structured Logging

Enable JSON event streaming for detailed feedback:

```hcl
navigator_mode = "json"
structured_logging = true
log_output_path = "./logs/build.json"
```

### 🔧 Enhanced Error Reporting

Get clear, actionable error messages:

```
ERROR: Play 'app.deploy' failed (exit code 2)
  └─ Task: "Install package"
  └─ Host: web-server-01
  └─ Check logs at: ./logs/build.json
```

## Common Use Cases

### Building Cloud Images

- [AWS AMI Building](EXAMPLES.md#aws-ec2-ami-building)
- [Azure VM Images](EXAMPLES.md#azure-vm-image)
- [Google Cloud Platform](EXAMPLES.md#google-cloud-platform)

### Container Images

- [Docker Images](EXAMPLES.md#building-docker-images)
- [Kubernetes-Ready Images](EXAMPLES.md#kubernetes-ready-images)

### Security and Compliance

- [CIS Hardened Images](EXAMPLES.md#cis-hardened-image)
- [HIPAA Compliant Infrastructure](EXAMPLES.md#hipaa-compliant-infrastructure)

### CI/CD Integration

- [GitHub Actions](EXAMPLES.md#github-actions)
- [GitLab CI](EXAMPLES.md#gitlab-ci)

## Version Compatibility

| Plugin Version | Packer Version | Ansible Navigator Version | Go Version |
|---------------|----------------|---------------------------|------------|
| 1.1.x | >= 1.7.0 | Any | >= 1.25.3 |
| 1.0.x | >= 1.7.0 | Any | >= 1.25.3 |

## Getting Help

### Support Channels

1. **Documentation**: You're here! Browse the guides above
2. **Issues**: [Report bugs](https://github.com/solomonhd/packer-plugin-ansible-navigator/issues)
3. **Discussions**: [Ask questions](https://github.com/solomonhd/packer-plugin-ansible-navigator/discussions)
4. **Examples**: [View examples](EXAMPLES.md)

### Before Reporting an Issue

1. Check the [Troubleshooting Guide](TROUBLESHOOTING.md)
2. Search existing [GitHub Issues](https://github.com/solomonhd/packer-plugin-ansible-navigator/issues)
3. Try the [Quick Diagnostics](TROUBLESHOOTING.md#quick-diagnostics)
4. Gather debug output with `PACKER_LOG=1`

## Contributing

We welcome contributions! See our [Contributing Guidelines](../CONTRIBUTING.md) for details.

### Development Resources

- [AGENTS.md](../AGENTS.md) - AI agent instructions
- [CHANGELOG.md](../CHANGELOG.md) - Version history
- [Development Setup](INSTALLATION.md#development-build) - Building from source

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](../LICENSE) file for details.

---

**Maintained by**: SolomonHD  
**Last Updated**: November 2025  
**Documentation Version**: 1.1.0
