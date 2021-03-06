// Copyright 2018 GRAIL, Inc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// The following enables go generate to generate the doc.go file.
//go:generate go run $GRAIL/go/src/vendor/v.io/x/lib/cmdline/testdata/gendoc.go "--build-cmd=go install" --copyright-notice= . -help
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	_ "github.com/grailbio/base/cmdutil/interactive"
	"github.com/grailbio/base/security/ticket"
	"v.io/v23/context"
	"v.io/v23/vdl"
	"v.io/x/lib/cmdline"
	"v.io/x/ref/lib/v23cmd"
	"v.io/x/ref/lib/vdl/codegen/json"
	_ "v.io/x/ref/runtime/factories/grail"
)

var (
	timeoutFlag       time.Duration
	authorityCertFlag string
	certFlag          string
	keyFlag           string
	jsonOnlyFlag      bool
)

func newCmdRoot() *cmdline.Command {
	root := &cmdline.Command{
		Runner: v23cmd.RunnerFunc(run),
		Name:   "grail-ticket",
		Short:  "Retrieve a ticket from a ticket-server",
		Long: `
Command grail-ticket retrieves a ticket from a ticket-server. A ticket is
identified using a Vanadium name.

Examples:

  grail ticket ticket/reflow/gdc/aws
  grail ticket /127.0.0.1:8000/reflow/gdc/aws

Note that tickets can be enumerated using the 'namespace' Vanadium tool:

  namespace glob /127.0.0.1:8000/...
  namespace glob /127.0.0.1:8000/reflow/...
`,
		ArgsName: "<ticket>",
		LookPath: false,
	}
	root.Flags.DurationVar(&timeoutFlag, "timeout", 10*time.Second, "Timeout for the requests to the ticket-server")
	root.Flags.BoolVar(&jsonOnlyFlag, "json-only", false, "Force a JSON output even for the tickets that have special handling")
	root.Flags.StringVar(&authorityCertFlag, "authority-cert", "", "PEM file to store the CA cert for a TLS-based ticket")
	root.Flags.StringVar(&certFlag, "cert", "", "PEM file to store the cert for a TLS-based ticket")
	root.Flags.StringVar(&keyFlag, "key", "", "PEM file to store the private key for a TLS-based ticket")
	return root
}

func saveCredentials(creds ticket.TlsCredentials) error {
	if err := ioutil.WriteFile(authorityCertFlag, []byte(creds.AuthorityCert), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(certFlag, []byte(creds.Cert), 0644); err != nil {
		return err
	}
	return ioutil.WriteFile(keyFlag, []byte(creds.Key), 0600)
}

func run(ctx *context.T, env *cmdline.Env, args []string) error {
	if len(args) != 1 {
		return env.UsageErrorf("Exactly one arguments (<ticket>) is required.")
	}

	ticketPath := args[0]
	client := ticket.TicketServiceClient(ticketPath)
	ctx, cancel := context.WithTimeout(ctx, timeoutFlag)
	defer cancel()

	t, err := client.Get(ctx)
	if err != nil {
		return err
	}

	jsonOutput := json.Const(vdl.ValueOf(t.Interface()), "", nil)

	if jsonOnlyFlag {
		fmt.Println(jsonOutput)
		return nil
	}

	if t.Index() == (ticket.TicketGenericTicket{}).Index() {
		fmt.Print(string((t.Interface().(ticket.GenericTicket)).Data))
		return nil
	}

	if len(authorityCertFlag)+len(certFlag)+len(keyFlag) > 0 {
		if len(authorityCertFlag)*len(certFlag)*len(keyFlag) == 0 {
			return fmt.Errorf("-authority-cert=%q, -cert=%q, -key=%q flags need to be all empty or all non-empty", authorityCertFlag, certFlag, keyFlag)
		}

		switch t.Index() {
		case (ticket.TicketDockerTicket{}).Index():
			return saveCredentials(t.(ticket.TicketDockerTicket).Value.Credentials)
		case (ticket.TicketDockerServerTicket{}).Index():
			return saveCredentials(t.(ticket.TicketDockerServerTicket).Value.Credentials)
		case (ticket.TicketDockerClientTicket{}).Index():
			return saveCredentials(t.(ticket.TicketDockerClientTicket).Value.Credentials)
		case (ticket.TicketTlsServerTicket{}).Index():
			return saveCredentials(t.(ticket.TicketTlsServerTicket).Value.Credentials)
		case (ticket.TicketTlsClientTicket{}).Index():
			return saveCredentials(t.(ticket.TicketTlsClientTicket).Value.Credentials)
		}
	}

	fmt.Println(jsonOutput)
	return nil
}

func main() {
	cmdline.HideGlobalFlagsExcept()
	cmdline.Main(newCmdRoot())
}
