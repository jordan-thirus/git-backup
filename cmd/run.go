/*
Copyright © 2024 Jordan Thirus <jordan.thirus@gmail.com>
*/
package cmd

import (
	"log/slog"

	"github.com/jordan-thirus/git-backup/internal/backup"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a one-time backup and archive",
	Run: func(cmd *cobra.Command, args []string) {
		Log.Debug("Run called")
		backup.Init(Cfg.Archive, Log)
		results := backup.Run(Cfg, Log)
		Log.Info("Run complete")
		Log.Debug("Results", slog.Any("results", results))
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
