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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

type Config struct {
	Nameservers []string
	Qtypes      []string
}

type LookupResult struct {
	DomainName string         `json:domainName`
	Result     []LookupRecord `json:results`
}

type LookupRecord struct {
	Nameserver string
	Records    []Record
}

type Record struct {
	Type   string
	Record []string
}

var config Config
var format string
var lres LookupResult

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
		setDomainName(args)
		switch format {
		case "stdout":
			lookupRecords(config, &lres)
			printStdout(lres)
		case "json":
			lookupRecords(config, &lres)
			printJSON(lres)
		default:
			lookupRecords(config, &lres)
			printStdout(lres)
		}
	},
}

func setDomainName(args []string) {
	lres.DomainName = args[0]
}

func printJSON(lres LookupResult) {
	json, _ := json.Marshal(lres)
	fmt.Println(string(json))
}

func printStdout(lres LookupResult) {
	fmt.Println(lres.DomainName)
	for _, lr := range lres.Result {
		fmt.Println(lr.Nameserver)
		for _, record := range lr.Records {
			for _, r := range record.Record {
				fmt.Println(record.Type + "\t" + r)
			}
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

func lookupRecords(config Config, lres *LookupResult) {
	for _, ns := range config.Nameservers {
		var records []Record
		r := getResolver(ns)
		for _, qtype := range config.Qtypes {
			lr := lookupRecord(lres.DomainName, qtype, r)
			records = append(records, Record{qtype, lr})
		}
		lres.Result = append(lres.Result, LookupRecord{"@" + ns, records})
	}
}

func lookupRecord(domainName string, qtype string, r *net.Resolver) []string {
	var result []string
	switch qtype {
	case "NS":
		res, err := r.LookupNS(context.Background(), domainName)
		if err != nil {
			return []string{"LookupNS error"}
		}
		for _, ns := range res {
			result = append(result, ns.Host)
		}
	case "A":
		res, err := r.LookupHost(context.Background(), domainName)
		if err != nil {
			return []string{"LookupHost error"}
		}
		for _, host := range res {
			result = append(result, host)
		}
	case "MX":
		res, err := r.LookupMX(context.Background(), domainName)
		if err != nil {
			return []string{"LookupMX error"}
		}
		for _, ns := range res {
			result = append(result, ns.Host)
		}
	}
	return result
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.diq.yaml)")

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
