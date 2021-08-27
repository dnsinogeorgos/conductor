package servicemanager

import (
	"context"
	"log"
	"sync"

	"github.com/coreos/go-systemd/v22/dbus"
)

type ServiceManager struct {
	mu              *sync.RWMutex
	MainServiceMane string
	Conn            *dbus.Conn
}

func NewServiceManager(name string) (*ServiceManager, error) {
	conn, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		return &ServiceManager{}, err
	}

	mu := &sync.RWMutex{}

	return &ServiceManager{mu, name, conn}, nil
}

func (mg *ServiceManager) StartMain() error {
	ctx := context.TODO()
	ch := make(chan string)

	jid, err := mg.Conn.StartUnitContext(ctx, mg.MainServiceMane, "fail", ch)
	if err != nil {
		return err
	}

	log.Printf("systemd job %d returned %s\n", jid, <-ch)

	return nil
}

func (mg *ServiceManager) StopMain() error {
	ctx := context.TODO()
	ch := make(chan string)

	jid, err := mg.Conn.StopUnitContext(ctx, mg.MainServiceMane, "fail", ch)
	if err != nil {
		return err
	}

	log.Printf("systemd job %d returned %s\n", jid, <-ch)

	return nil
}
