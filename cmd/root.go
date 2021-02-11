/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

type Config struct {
	Nameservers []string
	Query       []string
}

var config Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "diq",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domainName := args[0]
		fmt.Println((domainName))

		for _, ns := range config.Nameservers {
			fmt.Println("@" + ns)
			r := resolver(ns)
			for _, qs := range config.Query {
				lookupRecord(domainName, qs, r)
			}
			fmt.Println("")
		}
	},
}

func resolver(ns string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 10 * time.Second,
			}
			return d.DialContext(ctx, "udp", ns+":53")
		},
	}
}

func lookupRecord(domainName string, q string, r *net.Resolver) {
	switch q {
	case "NS":
		nss, err := r.LookupNS(context.Background(), domainName)
		if err != nil {
			fmt.Println("LookupNS error")
			os.Exit(1)
		}
		fmt.Print("  NS: ")
		for _, ns := range nss {
			//res = append(res, ns.Host)
			fmt.Println("  " + ns.Host)
		}
	case "A":
		hosts, err := r.LookupHost(context.Background(), domainName)
		if err != nil {
			fmt.Println("LookupHost error")
		}
		fmt.Print("A:")
		for _, host := range hosts {
			//res = append(res, host)
			fmt.Println("\t" + host)
		}
	case "MX":
		nss, err := r.LookupMX(context.Background(), domainName)
		if err != nil {
			fmt.Println("LookupMX error")
		}
		fmt.Print("MX:")
		for _, ns := range nss {
			//res = append(res, ns.Host)
			fmt.Println("\t" + ns.Host)
		}
	}
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.diq.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		/*home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}*/

		// Search config in home directory with name ".diq" (without extension).
		viper.SetConfigFile("~/$GOPATH/src/github.com/litencatt/diq/")
		viper.SetConfigName("diq")
	}

	//viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println(err)
	}
}
