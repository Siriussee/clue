package main

import (
	"encoding/json"
	"os"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/graph/graph"
	emu "github.com/anon/execution-property-graph/trace-emulator/emulator"
	"github.com/pkg/profile"
	"github.com/urfave/cli/v2"
)

const TRACE_FOLDER = "./trace/txs/"

func buildGraph(ctx *cli.Context) error {
	if ctx.Bool("profile") {
		defer profile.Start(profile.ProfilePath("."), profile.BlockProfile, profile.CPUProfile).Stop()
	}

	tx := ctx.String("tx")

	var trace emu.ExecutionResult
	var callFrame emu.CallFrame
	var header types.Header

	res, err := os.ReadFile(TRACE_FOLDER + tx + "_trace.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(res, &trace)
	if err != nil {
		return err
	}
	res, err = os.ReadFile(TRACE_FOLDER + tx + "_callframe.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(res, &callFrame)
	if err != nil {
		return err
	}
	res, err = os.ReadFile(TRACE_FOLDER + tx + "_header.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(res, &header)
	if err != nil {
		return err
	}

	txHash := types.Bytes32ToHash(types.HexToBytes32(tx))
	traceResult, err := graph.TraceOnEmulator(txHash, &trace, &callFrame, &header)
	if err != nil {
		return err
	}
	config := &graph.GraphConfig{
		RemoteUrl: ctx.String("remote"),
	}
	gb, err := graph.NewGraphBuilder(config)
	if err != nil {
		return err
	}
	g, err := gb.BuildGraph(traceResult)
	if err != nil {
		return err
	}
	g.Close()
	return nil
}
