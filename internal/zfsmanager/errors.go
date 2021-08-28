package zfsmanager

type LoadFilesystemChildrenError struct {
	s string
}

func (e LoadFilesystemChildrenError) Error() string {
	return e.s
}

type LoadCastStateError struct {
	s string
}

func (e LoadCastStateError) Error() string {
	return e.s
}

type LoadReplicaStateError struct {
	s string
}

func (e LoadReplicaStateError) Error() string {
	return e.s
}

type CastAlreadyExistsError struct {
	s string
}

func (e CastAlreadyExistsError) Error() string {
	return e.s
}

type CastNotFoundError struct {
	s string
}

func (e CastNotFoundError) Error() string {
	return e.s
}

type CastContainsReplicasError struct {
	s string
}

func (e CastContainsReplicasError) Error() string {
	return e.s
}

type ReplicaNotFoundError struct {
	s string
}

func (e ReplicaNotFoundError) Error() string {
	return e.s
}

type ReplicaAlreadyExistsError struct {
	s string
}

func (e ReplicaAlreadyExistsError) Error() string {
	return e.s
}

type StopMainError struct {
	s string
}

func (e StopMainError) Error() string {
	return e.s
}

type StartMainError struct {
	s string
}

func (e StartMainError) Error() string {
	return e.s
}
