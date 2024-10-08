/*
Copyright © 2024 Jordan Thirus <jordan.thirus@gmail.com>
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/jordan-thirus/git-backup/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var Cfg config.Configuration
var Log *slog.Logger

var rootCmd = &cobra.Command{
	Use:   "git-backup",
	Short: "Backup and archive git repositories",
	Long: `"git-backup" allows you to backup and/or archive git repostories
	on a configurable schedule.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLog, initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-backup.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".git-backup" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".git-backup")
	}

	viper.SetEnvPrefix("gb")
	viper.AutomaticEnv() // read in environment variables that match

	Cfg = config.New()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		viper.Unmarshal(&Cfg)
	}

	Log.Debug("configuration read", slog.Any("config", Cfg))
}

func initLog() {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}
