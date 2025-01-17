package main

import (
	"github.com/jrapoport/gothic/store"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:  "migrate",
	Long: "migrates the database",
	RunE: migrateRunE,
}

func migrateRunE(*cobra.Command, []string) error {
	c := rootConfig()
	l := c.Log().WithField("exe", "gmd")
	conn, err := store.Dial(c, l)
	if err != nil {
		return err
	}
	l.Info("starting migration...")
	err = conn.AutoMigrate()
	if err != nil {
		return err
	}
	l.Info("migration complete")
	return nil
}
