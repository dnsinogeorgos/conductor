package portmanager

import (
	"fmt"

	"go.uber.org/zap"
)

type PortManager struct {
	l          *zap.Logger
	LowerBound int32
	UpperBound int32
	PortMap    map[int32]string
}

func New(start int32, end int32, logger *zap.Logger) *PortManager {
	portMap := make(map[int32]string)
	pm := &PortManager{
		l:          logger,
		LowerBound: start,
		UpperBound: end,
		PortMap:    portMap,
	}

	logger.Sugar().Debugf("initialized portmanager with range %d to %d", start, end)

	return pm
}

func (pm *PortManager) Bind(port int32, name string) error {
	if n, found := pm.PortMap[port]; found {
		return PortInUseError{s: fmt.Sprintf("port %d is currently in use by %s\n", port, n)}
	}

	pm.l.Sugar().Debugf("binding name %s to port %d", name, port)
	pm.PortMap[port] = name

	return nil
}

func (pm *PortManager) Release(port int32) error {
	if _, found := pm.PortMap[port]; !found {
		return PortNotFoundError{s: fmt.Sprintf("port %d not found in list of bound ports\n", port)}
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

	return 0, PortsExhaustedError{s: fmt.Sprintf("available number of ports (%d) exhausted\n", len(portList))}
}

func (pm *PortManager) listAvailable() []int32 {
	pm.l.Sugar().Debugf("listing available ports starting from %d to %d", pm.LowerBound, pm.UpperBound)
	length := pm.UpperBound - pm.LowerBound + 1
	pm.l.Sugar().Debugf("creating an array of length %d", length)
	portList := make([]int32, length)
	fmt.Printf("dumping portList array: %v", portList)

	pm.l.Sugar().Debugf("iterating portList array")
	for i := range portList {
		portList[i] = pm.LowerBound + int32(i)
	}

	return portList
}
