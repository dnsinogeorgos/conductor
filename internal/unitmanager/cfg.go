package unitmanager

import (
	"bytes"
	"os"

	"go.uber.org/zap"
)

type unitConfig struct {
	Name    string
	Datadir string
	Port    int32
}

func (um *UnitManager) getConfigPath(cfg *unitConfig) (string, error) {
	var configPathBuffer bytes.Buffer

	err := um.configPathTemplate.Execute(&configPathBuffer, cfg)
	if err != nil {
		um.l.Error("could not render config file path", zap.Error(err))
		return "", err
	}

	return configPathBuffer.String(), nil
}

func (um *UnitManager) createConfig(name, datadir string, port int32) error {
	cfg := &unitConfig{
		Name:    name,
		Datadir: datadir,
		Port:    port,
	}

	cfgPath, err := um.getConfigPath(cfg)
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

func (um *UnitManager) cleanupConfig(name string) error {
	cfg := &unitConfig{
		Name: name,
	}

	cfgPath, err := um.getConfigPath(cfg)
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
