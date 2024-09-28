/*
Copyright © 2024 Jordan Thirus <jordan.thirus@gmail.com>
*/
package cmd

import (
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
		backup.Run(Cfg, Log)
		Log.Info("Run complete")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
