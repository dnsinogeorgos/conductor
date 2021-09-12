package unitmanager

import (
	"bytes"
	"os"

	"go.uber.org/zap"
)

// serviceConfig holds the systemd template unit variables
type serviceConfig struct {
	Name    string
	Datadir string
	Port    int32
}

// getServiceConfigPath returns the path of for the service configuration according to the
// configuration of the unit manager
func (um *UnitManager) getServiceConfigPath(cfg *serviceConfig) (string, error) {
	var configPathBuffer bytes.Buffer

	err := um.configPathTemplate.Execute(&configPathBuffer, cfg)
	if err != nil {
		um.l.Error("could not render config file path", zap.Error(err))
		return "", err
	}

	return configPathBuffer.String(), nil
}

// createServiceConfig creates the rendered service configuration file according to configuration
func (um *UnitManager) createServiceConfig(name, datadir string, port int32) error {
	cfg := &serviceConfig{
		Name:    name,
		Datadir: datadir,
		Port:    port,
	}

	cfgPath, err := um.getServiceConfigPath(cfg)
	if err != nil {
		return err
	}

	f, err := os.Create(cfgPath)
	if err != nil {
		um.l.Error("could not create cfg file on disk", zap.Error(err))
		return err
	}

	err = um.configFileTemplate.Execute(f, cfg)
	if err != nil {
		um.l.Error("could not render config file", zap.Error(err))
		return err
	}

	err = f.Close()
	if err != nil {
		um.l.Error("could not close cfg file", zap.Error(err))
		return err
	}

	return nil
}

// deleteServiceConfig cleans up the rendered service configuration file
func (um *UnitManager) deleteServiceConfig(name string) error {
	cfg := &serviceConfig{
		Name: name,
	}

	cfgPath, err := um.getServiceConfigPath(cfg)
	if err != nil {
		return err
	}

	err = os.Remove(cfgPath)
	if err != nil {
		um.l.Error("could not cleanup cfg file from disk", zap.Error(err))
		return err
	}

	return nil
}
