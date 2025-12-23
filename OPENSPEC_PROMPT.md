Add ssh_tunnel_mode option to the ansible-navigator provisioner that
automatically establishes an SSH tunnel through a bastion host to the
target, bypassing the Packer SSH proxy adapter. This enables reliable
connectivity when running execution environments in WSL2/Docker setups
where container-to-host networking is unreliable.
