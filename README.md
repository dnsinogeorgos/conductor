# Replica Conductor
Creates a ZFS pool on a block device, orchestrates an hierarchy of zfs datasets, list of
ports and systemd units.  
Includes tool to control the conductor service (conductorctl.py).

![](diagram.png)

### Nomenclature:
Pool -> the ZFS pool  
Filesystem -> The original filesystem that will be cloned to create casts  
Cast -> Transitory clone of the filesystem, anonymization/hooks will be applied here  
Replica -> Replica of the Cast

### Notes:
* The ZFS pool and the filesystem are initialized on the device if they do not exist.
* State of the casts and replicas is kept on the volumes, hence there is no need for
an external datastore. All casts and replicas will be loaded on start.

##### TODO: package to manage casts and pre/post hooks asynchronously from a goroutine

MariaDB 10.5 on ubuntu is showcased in vagrant. However, this is meant to be agnostic to
your database (or whatever you want to run with this).  
Conductor will manage for you:
- a block device and the zfs datasets
- a systemd unit (stop before clone, start after)
- a systemd template unit (start after creating replica, stop before deleting)
- configuration files generated from a provided template

### Quickstart:
Start vagrant `vagrant up` and ssh into the box `vagrant ssh`.  
An instance of MariaDB 10.5 will be started, with data in `/var/lib/mysql`. You can use
conductorctl and journalctl to see conductor in action.  

First, initialize mariadb so that you can actually use it.
```shell
cat configs/answers.txt | sudo mariadb-secure-install
```

Build and then run conductor with the provided configuration. Conductor logs access logs
to stdout, and json formatted application activity to stderr. Let's redirect stdout to
`/dev/null` to avoid the clutter.
```shell
go build -race cmd/conductor.go
sudo ./conductor -c configs/config.json 1>/dev/null
```

On a separate terminal, you can use conductorctl to create casts and replicas of the
running database instance.
```shell
conductorctl list # list existing replicas
conductorctl create -c example # create a cast named example
conductorctl create -c example -r john # create a replica of the example cast named john
```
As you may notice, the mariadb service is stopped right before snapshotting the main
dataset and started right back up. As the replica is created, a configuration file is
generated at `/etc/my.example_john.cnf` and the `mariadb@example_john.service` is
started.  
This is possible because the mariadb@.service template unit is configured and loaded
from `/etc/systemd/system/mariadb@.service`.

```shell
vagrant@ubuntu-focal:/vagrant$ conductorctl list
+----------------------+---------+---------+------+
|      Timestamp       |   Cast  | Replica | Port |
+----------------------+---------+---------+------+
| 2021-09-04T20:46:40Z | example |   john  | 3307 |
+----------------------+---------+---------+------+
vagrant@ubuntu-focal:/vagrant$ mysql -P 3307 -e 'status;'
--------------
mysql  Ver 15.1 Distrib 10.5.12-MariaDB, for debian-linux-gnu (x86_64) using readline 5.2

Connection id:          6
Current database:
Current user:           root@localhost
SSL:                    Not in use
Current pager:          stdout
Using outfile:          ''
Using delimiter:        ;
Server:                 MariaDB
Server version:         10.5.12-MariaDB-1:10.5.12+maria~focal mariadb.org binary distribution
Protocol version:       10
Connection:             127.0.0.1 via TCP/IP
Server characterset:    utf8mb4
Db     characterset:    utf8mb4
Client characterset:    utf8
Conn.  characterset:    utf8
TCP port:               3307
Uptime:                 1 min 12 sec

Threads: 1  Questions: 16  Slow queries: 0  Opens: 17  Open tables: 10  Queries per second avg: 0.222
--------------

vagrant@ubuntu-focal:/vagrant$
```

### Configuration

For this to work, you __must__ have a multi-service setup with systemd. More than enough
is provided in this example to get you started with MariaDB. You need to appropriately
edit your mariadb configuration and the replica configuration template. Ideally, the
main database instance is replicating from a remote source in order to decouple this
from your production systems. None of this is managed from conductor and never will be.

NOTE: to use this with real data and several replicas, you need to keep in mind the
resource requirements. In practice, with a ~250gb dataset and 4-5 replicas in use for
development and reporting usage, an instance with 64gb ram was used with an
`innodb_buffer_pool_size` value of 8gb for the main mariadb instance and for each of the
replicas.

Let's have a look at the configuration values available.

__debug__ is used for the zap logger. it lowers the log level and disables json
formatting  
__address__ is used for the router address string. default: `127.0.0.1`  
__port__ is used for the router address string. default: `8080`

__pool_name__ you can set the name of the zfs pool. default: `rootpool`  
__pool_path__ you can set the path of the zfs pool. default: `/rootpool`  
__pool_dev__ is the device conductor will create a zfs pool onto. *required*  
__filesystem_name__ is the name of the filesystem. default: `rootfs`  
__filesystem_path__ is the path where the main dataset will be mounted. the service
started from your main systemd unit must make use of this path. *required*  
__cast_path__ is the path where the casts will be mounted. default: `/rootfs_cast`  
__replica_path__ is the path where the replicas will be mounted. default:
`rootfs_replica`  

__port_from__ is the first port in the allocated range. these are not reserved in any
way, and you have to make sure this range will not be used by other processes.
*required*  
__port_to__ is the last port in the allocated range. *required*  
__main_unit__ is the main service unit name that will be managed. typically this will be
your main or replicating database. the dataset of this will be used for casts and
replicas. *required*  
__config_template_path__ is the template file that will be rendered for your service.
gotemplate syntax is used and available variables are `{{ .Name }}` `{{ .Datadir }}` and
`{{ .Port }}`. See `configs/myservice.cnd.tmpl` for a complete example. *required*  
__config_path_template_string__ is the path where the configuration template will be
rendered. gotemplate syntax is used and available variables are `{{ .Name }}` `{{ .Datadir }}` and
`{{ .Port }}`. an example of this is `/etc/my.{{ .Name }}.cnf`. *required*  
__unit_template_string__ is the systemd template unit that will be managed by conductor.
this unit must make use of the configuration files as configured with
`config_template_path` and `config_path_template_string`. *required*
