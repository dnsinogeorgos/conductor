### what now?

The goal is to use as much of the operating system instrumentation as possible. Starting
with Ubuntu and Debian, packaging for most OSes offers the wiring needed for running
multiple instances of MariaDB and possibly other databases as well.

NOTE: need to watch for possible SELinux and apparmor issues.

It is a challenge to keep this tool agnostic to the service that will run on top of the
filesystem.

Is a mariadb and a postgres example enough for this to be useable? Must provide
working example configurations and templates too.

Conductor starts and stops the systemd template units configured. e.g. mariadb results
in conductor managing the mariadb@3311.service unit.

Conductor manages the configuration using a provided template file. Also needs a path
and a filename. e.g. path = /etc/mysql/mariadb.conf.d and filename = 99-my$PORT.cnf. The
content of the file uses gotemplate syntax.

```
# /etc/mysql/mariadb.conf.d/99-multi.cnf
[mariadbd.3311]
port = 3311
socket = /run/mysqld/mysqld-3311.sock
datadir = /maria_replica/daily_20210824/carter

[mariadbd.{{ PORT }}]
port = {{ PORT }}
socket = /run/mysqld/mysqld-{{ PORT }}.sock
datadir = {{ DATADIR }}
```

```
# config values:
config_template_file
systemd_unit_name
generated_config_path
generated_config_filename

# template variables:
port
datadir
```

### What about anonymization hooks?
Conductor will receive an executable path (most likely a script) which will be executed
with the required environment variables (port and datadir). Webhooks could also be used.

```
# config values:
anonymization_hook_script
```
