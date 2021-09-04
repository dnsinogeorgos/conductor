package unitmanager

import (
	"context"
	"path"
	"text/template"

	"go.uber.org/zap"

	"github.com/coreos/go-systemd/v22/dbus"
)

type UnitManager struct {
	l                  *zap.Logger
	mainUnit           string
	unitTemplateString string
	configFileTemplate *template.Template
	unitNameTemplate   *template.Template
	configPathTemplate *template.Template
	conn               *dbus.Conn
}

func New(mu, ctp, uts, cpts string, logger *zap.Logger) *UnitManager {
	conn, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		logger.Fatal("could not connect to systemd", zap.Error(err))
		return &UnitManager{}
	}

	configFileTemplate, err := template.New(path.Base(ctp)).ParseFiles(ctp)
	if err != nil {
		logger.Fatal("could not load cfg template", zap.Error(err))
		return &UnitManager{}
	}

	unitNameTemplate, err := template.New("cfg").Parse(uts)
	if err != nil {
		logger.Fatal("could not load cfg template", zap.Error(err))
		return &UnitManager{}
	}

	configPathTemplate, err := template.New("cfg").Parse(cpts)
	if err != nil {
		logger.Fatal("could not load cfg template", zap.Error(err))
		return &UnitManager{}
	}

	unitmanager := &UnitManager{
		l:                  logger,
		mainUnit:           mu,
		configFileTemplate: configFileTemplate,
		unitNameTemplate:   unitNameTemplate,
		configPathTemplate: configPathTemplate,
		conn:               conn,
	}

	return unitmanager
}

func (um *UnitManager) StartMainUnit() error {
	ctx := context.TODO()
	ch := make(chan string)

	jid, err := um.conn.StartUnitContext(ctx, um.mainUnit, "fail", ch)
	if err != nil {
		um.l.Error("failed to start main unit", zap.Error(err))
	}

	um.l.Debug("systemd start main unit", zap.Int("job_id", jid), zap.String("value", <-ch))

	return nil
}

func (um *UnitManager) StopMainUnit() error {
	ctx := context.TODO()
	ch := make(chan string)

	jid, err := um.conn.StopUnitContext(ctx, um.mainUnit, "fail", ch)
	if err != nil {
		um.l.Error("failed to stop main unit", zap.Error(err))
	}

	um.l.Debug("systemd stop main unit", zap.Int("job_id", jid), zap.String("value", <-ch))

	return nil
}
