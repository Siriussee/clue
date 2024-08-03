package main

import "github.com/urfave/cli/v2"

var (
	ProfileFlag = &cli.BoolFlag{
		Name:  "profile",
		Usage: "start profile",
		Value: false,
	}
	TimePerformanceFlag = &cli.BoolFlag{
		Name:  "time",
		Usage: "time performance",
		Value: false,
	}
	GraphDBRemoteFlag = &cli.StringFlag{
		Name:  "remote",
		Usage: "graph database url",
		Value: "ws://127.0.0.1:8182/gremlin",
	}
	EthRemoteFlag = &cli.StringFlag{
		Name:    "eth-remote",
		Aliases: []string{"e"},
		Usage:   "eth node url (for headers and receipts)",
	}
	EthArchiveRemoteFlag = &cli.StringFlag{
		Name:    "eth-archive-remote",
		Aliases: []string{"ea"},
		Usage:   "eth archive node url",
	}
)
