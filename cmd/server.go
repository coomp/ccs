package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string
	listen  string
)

func NewCCSServerCmd() *cobra.Command {

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start CCS server with config or command line",
		Long:  `Start CCS server with config or command line.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting ccs server: " + viper.GetString("listen") + " ...")

		},
	}

	cobra.OnInitialize(initConfig)

	serverCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ccs.yaml)")
	serverCmd.PersistentFlags().StringVarP(&listen, "listen", "l", "localhost:2388", "tcp port for server")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	viper.BindPFlag("listen", serverCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("useViper", serverCmd.PersistentFlags().Lookup("viper"))

	return serverCmd
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ccs")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
