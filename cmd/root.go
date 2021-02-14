/*
Copyright © 2021 Kosuke Nakamura <ncl0709@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
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

	"github.com/litencatt/diq/version"
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
	Use:     "diq",
	Version: version.Version,
	Args:    cobra.MinimumNArgs(1),
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
		configFilePath := home + "/.diq.yml"
		file, err := os.OpenFile(configFilePath, os.O_RDONLY, 0)
		if err != nil {
			if os.IsNotExist(err) {
				file, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					fmt.Println(err)
				}
				defer file.Close()

				file.WriteString("nameservers:\n  - 8.8.8.8\nqtypes:\n  - A\n")
				fmt.Println("Create: " + configFilePath + "\n")
			}
		}
		defer file.Close()

		viper.SetConfigFile(configFilePath)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println(err)
	}
}
