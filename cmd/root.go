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

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Nameservers []string
	Qtypes      []string
}

type LookupResult struct {
	Domains []Domain `json:domains`
}

type Domain struct {
	DomainName string
	Result     []LookupRecord
}

type LookupRecord struct {
	Nameserver string
	Records    []Record
}

type Record struct {
	Type   string
	Record []string
}

var cfgFile string
var config Config
var format string
var qtype string
var lres LookupResult
var domainNames []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:  "diq",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setDomainNames(args)
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

func setDomainNames(args []string) {
	for _, domainName := range args {
		domainNames = append(domainNames, domainName)
	}
}

func printStdout(lres LookupResult) {
	for _, domain := range lres.Domains {
		fmt.Println(domain.DomainName)
		for _, lr := range domain.Result {
			fmt.Println(lr.Nameserver)
			for _, record := range lr.Records {
				for _, r := range record.Record {
					fmt.Println(record.Type + "\t" + r)
				}
			}
			fmt.Println("")
		}
	}
}

func printJSON(lres LookupResult) {
	json, _ := json.Marshal(lres)
	fmt.Println(string(json))
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
	for _, domainName := range domainNames {
		d := Domain{}
		d.DomainName = domainName
		for _, ns := range config.Nameservers {
			var records []Record
			r := getResolver(ns)
			for _, qtype := range getQtypes() {
				lr := lookupRecord(d.DomainName, qtype, r)
				records = append(records, Record{qtype, lr})
			}
			d.Result = append(d.Result, LookupRecord{"@" + ns, records})
		}
		lres.Domains = append(lres.Domains, d)
	}
}

func getQtypes() []string {
	if len(qtype) != 0 {
		return strings.Split(strings.ToUpper(qtype), ",")
	}
	return config.Qtypes
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
	case "TXT":
		res, err := r.LookupTXT(context.Background(), domainName)
		if err != nil {
			return []string{"LookupTXT error"}
		}
		for _, txt := range res {
			result = append(result, txt)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.diq.yml)")

	rootCmd.Flags().StringVarP(&format, "format", "f", "stdout", "output format. [stdout|json]")
	rootCmd.Flags().StringVarP(&qtype, "qtype", "q", "", "lookup query types. (e.g. -q a,mx)")
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

		viper.SetConfigFile(home + "/.diq.yml")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println(err)
	}
}
