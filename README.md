## deployd

deployd is a program I (plan to) use for deploying projects to different servers. It currently supports static websites and will support backend services written in Python and Go in the future.

deployd is opinionated - I will be using it for my projects on my servers, so it's designed to work with them.

## Configuration

deployd requires a configuration file to run. It'll look for one at `/etc/deployd.conf` but you can specify one with `--config`. The configuration file is YAML formatted.

### Sample Configuration File
(For reference when reading the usage).
```yaml
paths:
    static: /srv
    deployd: /srv/.deployd
static:
    - name: RHoK The Hood
      domain: rhokthehood.com
      subdomain: www
      branch: master
      bucket: rhokthehoodbuilds
    - name: Mihir Singh
      domain: mihirsingh.com
      subdomain: www
      branch: master
      github: True
      owner: citruspi
      repository: mihirsingh.com
```
