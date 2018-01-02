// Copyright © 2017 Nick Boughton
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"strconv"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/nboughton/dupes/file"
	"github.com/nboughton/go-utils/input"
	"github.com/nboughton/go-utils/regex/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

const (
	flFindOnly  = "find-only"
	flIgnoreDot = "ignore-dotfiles"
	flDir       = "dir"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dupes",
	Short: "recursively searches for duplicate files in a directory tree",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var (
			findOnly, _  = cmd.Flags().GetBool(flFindOnly)
			ignoreDot, _ = cmd.Flags().GetBool(flIgnoreDot)
			dir, _       = cmd.Flags().GetString(flDir)
		)

		if _, err := os.Stat(dir); err == os.ErrNotExist {
			fmt.Printf("No such directory [%s]\n", dir)
			return
		}

		fmt.Println("Preparing. This may take some time.")
		data, errors := file.ReadTree(dir, ignoreDot)
		for _, r := range data {
			if len(r.Paths) > 1 {
				fmt.Printf("\nDupes found for %s\n", r.Paths[0])
				for _, p := range r.Index() {
					fmt.Println("\t", p)
				}

				if findOnly {
					continue
				}

				fmt.Print("Remove dupes? [Y/n or index of file to keep]: ")
				ans := input.ReadLine()
				idx, err := strconv.Atoi(ans)
				if err != nil && !common.Yes.MatchString(ans) && ans != "" {
					fmt.Println("No action taken. Continuing.")
					continue
				}

				if err := r.Keep(idx); err != nil {
					errors = append(errors)
					fmt.Println(err)
					return
				}
			}
		}

		if len(errors) > 0 {
			fmt.Println("\nThe following errors occurred during the run:")
			for _, err := range errors {
				fmt.Println(err)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dupes.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.Flags().Bool(flFindOnly, false, "Toggles 'find only' mode. 'find only' lists dupes found but does not prompt for deletion")
	RootCmd.Flags().Bool(flIgnoreDot, false, "Toggles dotfile checking")
	RootCmd.Flags().StringP(flDir, "d", "", "Specify top level directory to scan.")
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

		// Search config in home directory with name ".dupes" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".dupes")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
