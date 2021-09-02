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

	logger.Info("initialized portmanager with range", zap.Int32("start_port", start), zap.Int32("end_port", end))

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
		pm.l.Fatal("incompatible configuration: tried to bind port outside of configured range", zap.Int32("port", port))
	}

	if n, found := pm.PortMap[port]; found {
		pm.l.Fatal("found inconsistent state: port is currently in use", zap.String("used_by", name), zap.Int32("port", port))
		return PortInUseError{p: port, n: n}
	}

	pm.l.Debug("binding name to port", zap.String("name", name), zap.Int32("port", port))
	pm.PortMap[port] = name

	return nil
}

func (pm *PortManager) Release(port int32) error {
	if _, found := pm.PortMap[port]; !found {
		pm.l.Fatal("found inconsistent state: port not found in list of used ports", zap.Int32("port", port))
		return PortNotFoundError{p: port}
	}

	pm.l.Debug("releasing port", zap.Int32("port", port))
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
