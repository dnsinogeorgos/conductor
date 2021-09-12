package unitmanager

import (
	"bytes"
	"context"
	"path"
	"text/template"

	"github.com/coreos/go-systemd/v22/dbus"
	"go.uber.org/zap"
)

// UnitManager manages the main systemd unit and the template units as configured
type UnitManager struct {
	l                  *zap.Logger
	mainUnit           string
	unitTemplateString string
	configFileTemplate *template.Template
	unitNameTemplate   *template.Template
	configPathTemplate *template.Template
	conn               *dbus.Conn
}

// New creates a UnitManager object
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

// StartMainUnit starts the configured main unit and returns error if unsuccessful
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

// StopMainUnit stops the configured main unit and returns error if unsuccessful
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

// StartTemplateUnit creates the related configuration file and starts the systemd template unit
// as configured
func (um *UnitManager) StartTemplateUnit(name, datadir string, port int32) error {
	ctx := context.TODO()
	ch := make(chan string)

	err := um.createServiceConfig(name, datadir, port)
	if err != nil {
		return err
	}

	unitName, err := um.getTemplateUnitName(name)
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

// StopTemplateUnit deletes the related configuration file and stops the systemd template unit
// unit as configured
func (um *UnitManager) StopTemplateUnit(name string) error {
	ctx := context.TODO()
	ch := make(chan string)

	unitName, err := um.getTemplateUnitName(name)
	if err != nil {
		return err
	}

	jid, err := um.conn.StopUnitContext(ctx, unitName, "fail", ch)
	if err != nil {
		um.l.Error("failed to stop unit", zap.String("name", name), zap.Error(err))
	}
	um.l.Debug("systemd stop unit", zap.String("unit", name), zap.Int("job_id", jid), zap.String("response", <-ch))

	err = um.deleteServiceConfig(name)
	if err != nil {
		return err
	}

	return nil
}

// getTemplateUnitName generates the full systemd template unit name according to configuration
func (um *UnitManager) getTemplateUnitName(name string) (string, error) {
	var unitNameBuffer bytes.Buffer

	err := um.unitNameTemplate.Execute(&unitNameBuffer, &struct{ Name string }{Name: name})
	if err != nil {
		um.l.Error("could not render unit name", zap.Error(err))
		return "", err
	}

	return unitNameBuffer.String(), nil
}
