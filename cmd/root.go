package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neomantra/tempted/internal/dev"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/////////////////////////////////////////////////////////////////////////////////////

type arg struct {
	cliShort, cliLong, cfgFileEnvVar, description string
}

var (
	version, sha string

	// Used for flags.
	cfgFile string

	cfgArg = arg{
		cliShort:    "c",
		cliLong:     "config",
		description: `Config file path. Default "$HOME/.tempted.yaml"`,
	}
	helpArg = arg{
		cliLong:     "help",
		description: `Print usage`,
	}
	addrArg = arg{
		cliShort:      "a",
		cliLong:       "address",
		cfgFileEnvVar: "temporal_cli_address",
		description:   `Nomad address. Default "localhost:7233"`,
	}
	namespaceArg = arg{
		cliShort:      "n",
		cliLong:       "namespace",
		cfgFileEnvVar: "temporal_namespace",
		description:   `Temporal namespace. Default "default"`,
	}
	updateSecondsArg = arg{
		cliShort:      "u",
		cliLong:       "update",
		cfgFileEnvVar: "tempted_update_seconds",
		description:   `Seconds between updates for workflow pages. Disable with "-1". Default "5"`,
	}
	logoColorArg = arg{
		cfgFileEnvVar: "tempted_logo_color",
	}

	description = `tempted is a terminal application for Temporal. It is used to
view workflows, and more, all from the terminal in a productivity-focused UI.`

	rootCmd = &cobra.Command{
		Use:     "tempted",
		Short:   "A terminal application for Temporal.",
		Long:    description,
		Run:     mainEntrypoint,
		Version: getVersion(),
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	// NOTE: default values here are unused even if default exists as they break the desired priority of cli args > env vars > config file > default if exists

	// root
	rootCmd.PersistentFlags().StringVarP(&cfgFile, cfgArg.cliLong, cfgArg.cliShort, "", cfgArg.description)
	rootCmd.PersistentFlags().BoolP(helpArg.cliLong, helpArg.cliShort, false, helpArg.description)
	for _, c := range []arg{
		addrArg,
		namespaceArg,
		updateSecondsArg,
	} {
		rootCmd.PersistentFlags().StringP(c.cliLong, c.cliShort, "", c.description)
		viper.BindPFlag(c.cliLong, rootCmd.PersistentFlags().Lookup(c.cfgFileEnvVar))
	}

	// colors, config or env var only
	viper.BindPFlag(logoColorArg.cliLong, rootCmd.PersistentFlags().Lookup(logoColorArg.cfgFileEnvVar))

	// serve
	for _, c := range []arg{
		hostArg,
		portArg,
		hostKeyPathArg,
		hostKeyPEMArg,
	} {
		serveCmd.PersistentFlags().StringP(c.cliLong, c.cliShort, "", c.description)
		viper.BindPFlag(c.cliLong, serveCmd.PersistentFlags().Lookup(c.cfgFileEnvVar))
	}

	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		_, err := os.Stat(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		extension := filepath.Ext(cfgFile)
		if extension != ".yaml" && extension != ".yml" {
			fmt.Println("error: config file must be .yaml or .yml")
			os.Exit(1)
		}

		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".tempted")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func mainEntrypoint(cmd *cobra.Command, args []string) {
	initialModel, options := setup(cmd, "")
	program := tea.NewProgram(initialModel, options...)

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on tempted startup: %v", err)
		os.Exit(1)
	}
}
