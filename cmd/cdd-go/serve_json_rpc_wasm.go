//go:build js || wasip1 || wasm

// Package main provides the CLI entrypoint and subcommand dispatcher for cdd-go.
package main

import (
	"flag"
	"fmt"
	"os"
)

// runServeJSONRPC parses CLI flags for the serve_json_rpc subcommand on WebAssembly targets where raw TCP HTTP listeners are unsupported.
// Parameter args is the slice of CLI arguments, and it returns any parse or unsupported platform error.
func runServeJSONRPC(args []string) error {
	fs := flag.NewFlagSet("serve_json_rpc", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var port int
	var listen string

	fs.IntVar(&port, "port", envOrDefaultInt("CDD_PORT", 8080), "Port to listen on")
	fs.IntVar(&port, "p", envOrDefaultInt("CDD_PORT", 8080), "Port to listen on (shorthand)")
	fs.StringVar(&listen, "listen", envOrDefault("CDD_LISTEN", "127.0.0.1"), "Address to listen on")
	fs.StringVar(&listen, "l", envOrDefault("CDD_LISTEN", "127.0.0.1"), "Address to listen on (shorthand)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	return fmt.Errorf("serve_json_rpc is not supported in WebAssembly builds; use 'mcp' over stdio instead")
}
