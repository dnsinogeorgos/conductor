package unitmanager

import (
	"context"

	"go.uber.org/zap"
)

func (um *UnitManager) StartUnit(name string) error {
	ctx := context.TODO()
	ch := make(chan string)

	jid, err := um.Conn.StartUnitContext(ctx, name, "fail", ch)
	if err != nil {
		um.l.Error("failed to start unit", zap.String("name", name), zap.Error(err))
	}

	um.l.Debug("systemd start unit", zap.String("unit", name), zap.Int("job_id", jid), zap.String("value", <-ch))

	return nil
}

func (um *UnitManager) StopUnit(name string) error {
	ctx := context.TODO()
	ch := make(chan string)

	jid, err := um.Conn.StopUnitContext(ctx, name, "fail", ch)
	if err != nil {
		um.l.Error("failed to stop unit", zap.String("name", name), zap.Error(err))
	}

	um.l.Debug("systemd stop unit", zap.String("unit", name), zap.Int("job_id", jid), zap.String("value", <-ch))

	return nil
}
