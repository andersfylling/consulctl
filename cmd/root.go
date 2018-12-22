// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var consulAddress string
var consulPort string
var protocol string
var version bool

var baseURL string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "consulctl",
	Short: "For interacting with consul through your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			fmt.Println("consulctl v0.1.0")
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.consulctl.yaml)")
	rootCmd.PersistentFlags().StringVar(&consulAddress, "consul-address", "", "the ip of the consul agent for API communication")
	rootCmd.PersistentFlags().StringVar(&consulPort, "consul-port", "", "the port of the consul agent for API communication")
	rootCmd.PersistentFlags().StringVar(&protocol, "protocol", "", "'http' or 'https' when contacting the consul agent over it's API")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "consulctl version")


}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".consulctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".consulctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}


	// setup base url for API access

	// I acknowledge it can be difficult to add it as a hostname, so "consul-node" is the fallback if
	// no env-vars have been defined and the --consul-address is not set
	if consulAddress == "" {
		consulAddress = os.Getenv("CONSULE_NODE")
		if consulAddress == "" {
			consulAddress = "consul-node"
		}
	}

	if protocol == "" {
		protocol = "http"
	}

	if consulPort == "" {
		consulPort = "8500"
	}

	baseURL = protocol + "://" + consulAddress + ":" + consulPort + "/v1"
}
