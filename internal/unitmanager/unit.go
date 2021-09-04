package unitmanager

import (
	"bytes"
	"context"

	"go.uber.org/zap"
)

func (um *UnitManager) getUnitName(name string) (string, error) {
	var unitNameBuffer bytes.Buffer

	err := um.unitNameTemplate.Execute(&unitNameBuffer, &struct{ Name string }{Name: name})
	if err != nil {
		um.l.Error("could not render unit name", zap.Error(err))
		return "", err
	}

	return unitNameBuffer.String(), nil
}

func (um *UnitManager) StartUnit(name, datadir string, port int32) error {
	ctx := context.TODO()
	ch := make(chan string)

	err := um.createConfig(name, datadir, port)
	if err != nil {
		return err
	}

	unitName, err := um.getUnitName(name)
	if err != nil {
		return err
	}

	jid, err := um.conn.StartUnitContext(ctx, unitName, "fail", ch)
	if err != nil {
		um.l.Error("failed to start unit", zap.String("name", name), zap.Error(err))
	}
	um.l.Debug("systemd start unit", zap.String("unit", name), zap.Int("job_id", jid), zap.String("response", <-ch))

	return nil
}

func (um *UnitManager) StopUnit(name string) error {
	ctx := context.TODO()
	ch := make(chan string)

	unitName, err := um.getUnitName(name)
	if err != nil {
		return err
	}

	jid, err := um.conn.StopUnitContext(ctx, unitName, "fail", ch)
	if err != nil {
		um.l.Error("failed to stop unit", zap.String("name", name), zap.Error(err))
	}
	um.l.Debug("systemd stop unit", zap.String("unit", name), zap.Int("job_id", jid), zap.String("response", <-ch))

	err = um.cleanupConfig(name)
	if err != nil {
		return err
	}

	return nil
}
