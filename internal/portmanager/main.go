package portmanager

import (
	"go.uber.org/zap"
)

type PortManager struct {
	l          *zap.Logger
	LowerBound int32
	UpperBound int32
	PortMap    map[int32]string
}

func New(start int32, end int32, logger *zap.Logger) *PortManager {
	if end < start {
		logger.Fatal("bad configuration: end port cannot be lower than start port")
	}

	if start == 0 {
		logger.Fatal("bad configuration: start port cannot be 0")
	}

	portMap := make(map[int32]string)
	pm := &PortManager{
		l:          logger,
		LowerBound: start,
		UpperBound: end,
		PortMap:    portMap,
	}

	logger.Sugar().Infof("initialized portmanager with range %d to %d", start, end)

	return pm
}

func (pm *PortManager) Bind(port int32, name string) error {
	isValid := false

	portList := pm.listAvailable()
	for _, tryPort := range portList {
		if port == tryPort {
			isValid = true
			break
		}
	}
	if isValid == false {
		pm.l.Sugar().Fatalf("incompatible configuration: tried to bind port %d which is outside of the configured range", port)
	}

	if n, found := pm.PortMap[port]; found {
		pm.l.Sugar().Fatalf("found inconsistent state: port %d is currently in use by %s", port, n)
		return PortInUseError{p: port, n: n}
	}

	pm.l.Sugar().Debugf("binding name %s to port %d", name, port)
	pm.PortMap[port] = name

	return nil
}

func (pm *PortManager) Release(port int32) error {
	if _, found := pm.PortMap[port]; !found {
		pm.l.Sugar().Fatalf("found inconsistent state: port %d not found in list of used ports", port)
		return PortNotFoundError{p: port}
	}

	pm.l.Sugar().Debugf("releasing port %d", port)
	delete(pm.PortMap, port)
	return nil
}

func (pm *PortManager) GetNextAvailable() (int32, error) {
	portList := pm.listAvailable()
	for _, port := range portList {
		if _, found := pm.PortMap[port]; !found {
			return port, nil
		}
	}

	return 0, PortsExhaustedError{}
}

func (pm *PortManager) listAvailable() []int32 {
	length := pm.UpperBound - pm.LowerBound + 1
	portList := make([]int32, length)

	for i := range portList {
		portList[i] = pm.LowerBound + int32(i)
	}

	return portList
}
