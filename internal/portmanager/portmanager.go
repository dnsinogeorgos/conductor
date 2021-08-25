package portmanager

import (
	"fmt"
	"log"
)

type PortManager struct {
	LowerBound uint16
	UpperBound uint16
	PortMap    map[uint16]string
}

func NewPortManager(start uint16, end uint16) *PortManager {
	portMap := make(map[uint16]string)
	pm := &PortManager{
		LowerBound: start,
		UpperBound: end,
		PortMap:    portMap,
	}

	log.Printf("using ports %d to %d\n", start, end)
	return pm
}

func (pm *PortManager) Bind(port uint16, name string) error {
	if n, found := pm.PortMap[port]; found {
		return PortInUseError{s: fmt.Sprintf("port %d is currently in use by %s\n", port, n)}
	}

	log.Printf("binding name %s to port %d\n", name, port)
	pm.PortMap[port] = name

	return nil
}

func (pm *PortManager) Release(port uint16) error {
	if _, found := pm.PortMap[port]; !found {
		return PortNotFoundError{s: fmt.Sprintf("port %d not found in list of bound ports\n", port)}
	}

	log.Printf("releasing port %d\n", port)
	delete(pm.PortMap, port)
	return nil
}

func (pm *PortManager) GetNextAvailable() (uint16, error) {
	portList := pm.listAvailable()
	for _, port := range portList {
		if _, found := pm.PortMap[port]; !found {
			return port, nil
		}
	}

	return 0, PortsExhaustedError{s: fmt.Sprintf("available number of ports (%d) exhausted\n", len(portList))}
}

func (pm *PortManager) listAvailable() []uint16 {
	length := pm.UpperBound - pm.LowerBound + 1
	portList := make([]uint16, length)

	for i := range portList {
		portList[i] = pm.LowerBound + uint16(i)
	}

	return portList
}
