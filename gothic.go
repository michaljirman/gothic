package gothic

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jrapoport/gothic/config"
	"github.com/jrapoport/gothic/core"
	"github.com/jrapoport/gothic/hosts"
	"github.com/segmentio/encoding/json"
)

// Main is the application main
func Main(c *config.Config) error {
	if c.IsDebug() {
		b, _ := json.MarshalIndent(c, "", "\t")
		fmt.Println(string(b))
	}
	signalsToCatch := []os.Signal{
		os.Interrupt,
		os.Kill,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGABRT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, signalsToCatch...)
	a, err := core.NewAPI(c)
	if err != nil {
		return err
	}
	defer func() {
		if err = a.Shutdown(); err != nil {
			c.Log().Error(err)
			return
		}
		c.Log().Infof("%s shut down", c.Name)
	}()
	c.Log().Infof("starting %s...", c.Name)
	err = hosts.Start(a, c)
	if err != nil {
		return err
	}
	defer func() {
		if err = hosts.Shutdown(); err != nil {
			c.Log().Error(err)
			return
		}
		c.Log().Infof("%s shut down", c.Name)
	}()
	c.Log().Infof("%s %s started", c.Name, c.Version())
	<-stopCh
	c.Log().Infof("%s shutting down", c.Name)
	return nil
}
