Plan out a golang tool that acts as a test/verification tool. I want the system to be extensible and plugin-able. I
ant to be able to test aws infrastructure, openstack infrastructure, linux packages, linux filesystems, etc. I want
to start small right now with linux testing via ssh.

Example usage:

```
./platform-spec test ssh -i <path to private key> ubuntu@<IP> ./test.yaml
```

The tests should look something like

```
# test.yaml

packages:
  - docker-ce
  - docker-compose
directory:
  - /opt
  - /opt/monitoring
file:
  - /opt/monitoring/docker-compose.yml
line-in-file:
  - /opt/monitoring/docker-compose.yml:"registry.${DOMAIN}/docker-hub/grafana/grafana:12.2.1"
```
