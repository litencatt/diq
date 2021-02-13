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
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

type Config struct {
	Nameservers []string
	Query       []string
}

type LookupResult struct {
	DomainName string         `json:domainName`
	Result     []LookupRecord `json:results`
}

type LookupRecord struct {
	Nameserver string
	Records    []string
}

var config Config
var format string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "diq",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domainName := args[0]
		lrs := LookupResult{}
		lrs.DomainName = domainName
		for _, ns := range config.Nameservers {
			var records []string
			r := getResolver(ns)
			for _, query := range config.Query {
				records = append(records, lookupRecord(domainName, query, r))
			}
			lrs.Result = append(lrs.Result, LookupRecord{"@" + ns, records})
		}

		switch format {
		case "stdout":
			printResults(lrs)
		case "json":
			printJSON(lrs)
		default:
			printResults(lrs)
		}
	},
}

func printJSON(lrs LookupResult) {
	json, _ := json.Marshal(lrs)
	fmt.Println(string(json))
}

func printResults(lrs LookupResult) {
	fmt.Println(lrs.DomainName)
	for _, lr := range lrs.Result {
		fmt.Println(lr.Nameserver)
		for _, record := range lr.Records {
			fmt.Println(record)
		}
		fmt.Println("")
	}
}

func getResolver(ns string) *net.Resolver {
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

func lookupRecord(domainName string, q string, r *net.Resolver) string {
	var result []string
	switch q {
	case "NS":
		res, err := r.LookupNS(context.Background(), domainName)
		if err != nil {
			return "LookupNS error"
		}
		for _, ns := range res {
			result = append(result, "NS:\t"+ns.Host)
		}
	case "A":
		res, err := r.LookupHost(context.Background(), domainName)
		if err != nil {
			return "LookupHost error"
		}
		for _, host := range res {
			result = append(result, "A\t"+host)
		}
	case "MX":
		res, err := r.LookupMX(context.Background(), domainName)
		if err != nil {
			return "LookupMX error"
		}
		for _, ns := range res {
			result = append(result, "MX\t"+ns.Host)
		}
	}
	return strings.Join(result, ",")
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
	rootCmd.Flags().StringVarP(&format, "format", "f", "stdout", "output format. [stdout|json]")
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
