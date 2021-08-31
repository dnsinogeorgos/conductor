### api
this package provides the router object. it requires a conductor object and passes requests to
the appropriate conductor methods. also communicates state to the caller.

### conductor
this is the main package that contains all the logic. it wires configuration, managing
port numbers, managing the systemd units and the moulder. it is called through the api
module.

### config
this package prepares the config object.

### portmanager
this package creates the portmanager. it keeps track of port assignment for
casts and replicas.

### TODO: servicemanager
this package creates the servicemanager object. it keeps track of systemd units.
it manages the main unit, the template units and generates configuration from templates.
this will run concurrently and receive commands asynchronously through channels. proper
backoff timers will be applied to avoid flapping the main unit. will provide appropriate
logging to not cause confusion.

### zfsmanager
this package manages the zfs datasets. this includes pool, filesystem, snapshotting and
cloning of the filesystem. also manages and communicates some state that is persisted
onto the dataset.

### moulder
this package will implement the cast lifecycle. it will be called from the create cast
method. for this, the cast state will be updated with a status value. during the
preparing phase, a template unit will be started for the lifetime of the hook execution.

---

- [x] logging
  - [x] stdout for access logs, stderr for the rest
  - [x] use minimal log levels, just debug and warn
  - [ ] switch to cheney errors with interfaces instead of types(?)
- [x] api
- [ ] zfs
  - [x] support config from json but also environment variables
  - [ ] cast lifecycle and status in state: created -> preparing -> ready
  - [ ] implement cast creation hooks
  - [ ] support different zfs setups
- [ ] portmanager
  - [ ] add proper code comments
- [ ] servicemanager
  - [x] main service stop/start
  - [ ] replica service start/stop
  - [ ] cast service start/stop for hooks
  - [ ] run as goroutine and communicate over channels
  - [ ] add proper rate limits / backoff timers
- [ ] other
  - [x] cleanup main, create run
  - [ ] handle signals, exit gracefully after getting lock
  - [ ] write tests
