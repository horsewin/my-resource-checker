package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool
var outputFormat string
var region string
var profile string

var rootCmd = &cobra.Command{
	Use:   "sbcntr-validator",
	Short: "AWS Resource Validator for SBCNTR Hands-on",
	Long: `AWS Resource Validator is a CLI tool that validates AWS resources
created during the SBCNTR hands-on exercises using AWS Cloud Control API.

It checks if resources are correctly created and configured at each step,
helping learners ensure they are progressing correctly.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sbcntr-validator.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "console", "output format (console, json)")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "ap-northeast-1", "AWS region")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "AWS profile")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".sbcntr-validator")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
