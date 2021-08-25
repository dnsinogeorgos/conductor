package systemd

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

func MainService(ctx context.Context, name string) error {

	svc, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	svcCh := make(chan string)
	jid, err := svc.StartUnitContext(ctx, name, "fail", svcCh)
	if err != nil {
		return err
	}

	fmt.Printf("systemd job %d returned %s\n", jid, <-svcCh)

	return nil
}
