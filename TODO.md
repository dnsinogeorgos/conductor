### api
this package provides the router object. it requires a zfs object and passes requests to
the appropriate zfs methods.

### zfs
this is the main package that contains all the logic. it wires configuration, managing
port numbers, managing the systemd units and the moulder. it is called through the api
module.

### config
this package holds the config objects. has a reason to exist and will be moved into zfs.
currently has dependencies on zfs, portmanager and service manager.

### portmanager
this package creates the portmanager object. it keeps track of port assignment for
casts and replicas. has a proper set of errors and is generally pretty clean.

### servicemanager
this package creates the servicemanager object. it keeps track of systemd units.
it manages the main unit, the template units and generates configuration from templates.
this will run concurrently and receive commands asynchronously through channels. proper
backoff timers will be applied to avoid flapping the main unit. will provide appropriate
logging to not cause confusion. this will also run concurrently and receive commands
asynchronously through channels.

### moulder
this package will implement the cast lifecycle. it will be called from the create cast
method. for this, the cast state will be updated with a status value. during the
preparing phase, a template unit will be started for the lifetime of the hook execution.

---

- [x] api
- [ ] zfs
  - [ ] move config structs into zfs module
  - [ ] cast lifecycle created -> preparing -> ready
  - [ ] implement cast creation hooks
  - [ ] support different zfs setups
- [ ] portmanager
  - [ ] add proper code comments
- [ ] servicemanager
  - [x] implement main service stop/start
  - [ ] implement replica service start/stop
  - [ ] implement cast service start/stop
