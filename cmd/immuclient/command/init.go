/*
Copyright 2019-2020 vChain, Inc.

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

package immuclient

import (
	"fmt"
	"os"

	"github.com/codenotary/immudb/cmd/docs/man"
	c "github.com/codenotary/immudb/cmd/helper"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/gw"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type commandline struct {
	ImmuClient     client.ImmuClient
	passwordReader c.PasswordReader
	valueOnly      bool
	options        *client.Options
}

func Init(cmd *cobra.Command, o *c.Options) {
	if err := configureOptions(cmd, o); err != nil {
		c.QuitToStdErr(err)
	}
	cl := new(commandline)
	cl.passwordReader = c.DefaultPasswordReader
	cl.options = options()
	// login and logout
	cl.login(cmd)
	cl.logout(cmd)
	// current status
	cl.currentRoot(cmd)
	// get operations
	cl.getByIndex(cmd)
	cl.getKey(cmd)
	cl.rawSafeGetKey(cmd)
	cl.safeGetKey(cmd)
	// set operations
	cl.rawSafeSet(cmd)
	cl.set(cmd)
	cl.safeset(cmd)
	cl.zAdd(cmd)
	cl.safeZAdd(cmd)
	// scanners
	cl.zScan(cmd)
	cl.iScan(cmd)
	cl.scan(cmd)
	cl.count(cmd)
	// references
	cl.reference(cmd)
	cl.safereference(cmd)
	// misc
	cl.inclusion(cmd)
	cl.consistency(cmd)
	cl.history(cmd)
	cl.status(cmd)
	// man file generator
	cmd.AddCommand(man.Generate(cmd, "immuclient", "./cmd/docs/man/immuclient"))
}

func options() *client.Options {
	port := viper.GetInt("immudb-port")
	address := viper.GetString("immudb-address")
	tokenFileName := viper.GetString("tokenfile")
	mtls := viper.GetBool("mtls")
	certificate := viper.GetString("certificate")
	servername := viper.GetString("servername")
	pkey := viper.GetString("pkey")
	clientcas := viper.GetString("clientcas")
	options := client.DefaultOptions().
		WithPort(port).
		WithAddress(address).
		WithTokenFileName(tokenFileName).
		WithMTLs(mtls)
	if mtls {
		// todo https://golang.org/src/crypto/x509/root_linux.go
		options.MTLsOptions = client.DefaultMTLsOptions().
			WithServername(servername).
			WithCertificate(certificate).
			WithPkey(pkey).
			WithClientCAs(clientcas)
	}
	return options
}

func (cl *commandline) connect(cmd *cobra.Command, args []string) (err error) {
	opts := options()
	len, err := client.ReadFileFromUserHomeDir(opts.TokenFileName)
	if err != nil || len == "" {
		opts.Auth = false
	} else {
		opts.Auth = true
	}
	if cl.ImmuClient, err = client.NewImmuClient(opts); err != nil || cl.ImmuClient == nil {
		c.QuitToStdErr(err)
	}
	cl.valueOnly = viper.GetBool("value-only")
	return
}

func (cl *commandline) disconnect(cmd *cobra.Command, args []string) {
	if err := cl.ImmuClient.Disconnect(); err != nil {
		c.QuitToStdErr(err)
	}
	os.Exit(0)
}

func configureOptions(cmd *cobra.Command, o *c.Options) error {
	cmd.PersistentFlags().IntP("immudb-port", "p", gw.DefaultOptions().ImmudbPort, "immudb port number")
	cmd.PersistentFlags().StringP("immudb-address", "a", gw.DefaultOptions().ImmudbAddress, "immudb host address")
	cmd.PersistentFlags().StringVar(&o.CfgFn, "config", "", "config file (default path are configs or $HOME. Default filename is immuclient.toml)")
	cmd.PersistentFlags().String(
		"tokenfile",
		client.DefaultOptions().TokenFileName,
		fmt.Sprintf(
			"authentication token file (default path is $HOME or binary location; default filename is %s)",
			client.DefaultOptions().TokenFileName))
	cmd.PersistentFlags().BoolP("mtls", "m", client.DefaultOptions().MTLs, "enable mutual tls")
	cmd.PersistentFlags().String("servername", client.DefaultMTLsOptions().Servername, "used to verify the hostname on the returned certificates")
	cmd.PersistentFlags().String("certificate", client.DefaultMTLsOptions().Certificate, "server certificate file path")
	cmd.PersistentFlags().String("pkey", client.DefaultMTLsOptions().Pkey, "server private key path")
	cmd.PersistentFlags().String("clientcas", client.DefaultMTLsOptions().ClientCAs, "clients certificates list. Aka certificate authority")
	cmd.PersistentFlags().Bool("value-only", false, "returning only values for get operations")
	if err := viper.BindPFlag("immudb-port", cmd.PersistentFlags().Lookup("immudb-port")); err != nil {
		return err
	}
	if err := viper.BindPFlag("immudb-address", cmd.PersistentFlags().Lookup("immudb-address")); err != nil {
		return err
	}
	if err := viper.BindPFlag("auth", cmd.PersistentFlags().Lookup("auth")); err != nil {
		return err
	}
	if err := viper.BindPFlag("tokenfile", cmd.PersistentFlags().Lookup("tokenfile")); err != nil {
		return err
	}
	if err := viper.BindPFlag("mtls", cmd.PersistentFlags().Lookup("mtls")); err != nil {
		return err
	}
	if err := viper.BindPFlag("servername", cmd.PersistentFlags().Lookup("servername")); err != nil {
		return err
	}
	if err := viper.BindPFlag("certificate", cmd.PersistentFlags().Lookup("certificate")); err != nil {
		return err
	}
	if err := viper.BindPFlag("pkey", cmd.PersistentFlags().Lookup("pkey")); err != nil {
		return err
	}
	if err := viper.BindPFlag("clientcas", cmd.PersistentFlags().Lookup("clientcas")); err != nil {
		return err
	}
	if err := viper.BindPFlag("value-only", cmd.PersistentFlags().Lookup("value-only")); err != nil {
		return err
	}

	viper.SetDefault("immudb-port", gw.DefaultOptions().ImmudbPort)
	viper.SetDefault("immudb-address", gw.DefaultOptions().ImmudbAddress)
	viper.SetDefault("auth", client.DefaultOptions().Auth)
	viper.SetDefault("tokenfile", client.DefaultOptions().TokenFileName)
	viper.SetDefault("mtls", client.DefaultOptions().MTLs)
	viper.SetDefault("servername", client.DefaultMTLsOptions().Servername)
	viper.SetDefault("certificate", client.DefaultMTLsOptions().Certificate)
	viper.SetDefault("pkey", client.DefaultMTLsOptions().Pkey)
	viper.SetDefault("clientcas", client.DefaultMTLsOptions().ClientCAs)
	viper.SetDefault("value-only", false)
	return nil
}
