package unitmanager

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/coreos/go-systemd/v22/dbus"
)

type UnitManager struct {
	mu              *sync.RWMutex
	l               *zap.Logger
	MainServiceName string
	Conn            *dbus.Conn
}

func New(name string, logger *zap.Logger) *UnitManager {
	conn, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		logger.Fatal("could not connect to systemd", zap.Error(err))
	}

	mu := &sync.RWMutex{}
	l := logger

	return &UnitManager{mu, l, name, conn}
}

func (um *UnitManager) StartMainUnit() error {
	ctx := context.TODO()
	ch := make(chan string)

	jid, err := um.Conn.StartUnitContext(ctx, um.MainServiceName, "fail", ch)
	if err != nil {
		um.l.Error("failed to start main unit", zap.Error(err))
	}

	um.l.Sugar().Debugf("systemd job %d returned %s", jid, <-ch)

	return nil
}

func (um *UnitManager) StopMainUnit() error {
	ctx := context.TODO()
	ch := make(chan string)

	jid, err := um.Conn.StopUnitContext(ctx, um.MainServiceName, "fail", ch)
	if err != nil {
		um.l.Error("failed to stop main unit", zap.Error(err))
	}

	um.l.Sugar().Debugf("systemd job %d returned %s", jid, <-ch)

	return nil
}
