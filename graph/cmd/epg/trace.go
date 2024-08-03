package main

import (
	"context"
	"encoding/json"
	"math/big"
	"os"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	emu "github.com/anon/execution-property-graph/trace-emulator/emulator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/urfave/cli/v2"
)

func trace(txHash common.Hash, rpcClient *rpc.Client, receiptClient *ethclient.Client) (*emu.ExecutionResult, *emu.CallFrame, *types.Header, error) {
	var callFrame emu.CallFrame

	callFrameOptions := map[string]interface{}{
		"tracer": "callTracer",
	}

	if err := rpcClient.Call(&callFrame, "debug_traceTransaction", txHash.Hex(), callFrameOptions); err != nil {
		log.Error("debug_traceTransaction using callframe tracer error for tx ", txHash.Hex(), " err: ", err)
		return nil, nil, nil, err
	}

	var trace emu.ExecutionResult

	structLogOptions := map[string]interface{}{
		"enableMemory":     true,
		"disableStack":     false,
		"disableStorage":   false,
		"enableReturnData": true,
	}

	if err := rpcClient.Call(&trace, "debug_traceTransaction", txHash.Hex(), structLogOptions); err != nil {
		log.Error("debug_traceTransaction using structlogs tracer error for tx ", txHash.Hex(), " err: ", err)
		return nil, nil, nil, err
	}

	if receiptClient == nil {
		receiptClient = ethclient.NewClient(rpcClient)
	}

	receipt, err := receiptClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Warn("Failed to fetch transaction receipt: ", err)
		return nil, nil, nil, err
	}

	ethHeader, err := receiptClient.HeaderByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		log.Warn("Failed to fetch block header: ", err)
		return nil, nil, nil, err
	}

	header := &types.Header{
		Number:     new(big.Int).Set(ethHeader.Number),
		Difficulty: new(big.Int).Set(ethHeader.Difficulty),
	}

	return &trace, &callFrame, header, nil
}

func serialize(trace *emu.ExecutionResult, callFrame *emu.CallFrame, header *types.Header, txHash common.Hash) error {
	traceJson, err := json.Marshal(trace)
	if err != nil {
		log.Error("marshal trace error for tx ", txHash.Hex(), " err: ", err)
		return err
	}
	os.WriteFile(TRACE_FOLDER+txHash.Hex()+"_trace.json", traceJson, 0644)

	callFrameJson, err := json.Marshal(callFrame)
	if err != nil {
		log.Error("marshal call frame error for tx ", txHash.Hex(), " err: ", err)
		return err
	}
	os.WriteFile(TRACE_FOLDER+txHash.Hex()+"_callframe.json", callFrameJson, 0644)

	headerJson, err := json.Marshal(header)
	if err != nil {
		log.Error("marshal header error for tx ", txHash.Hex(), " err: ", err)
		return err
	}
	os.WriteFile(TRACE_FOLDER+txHash.Hex()+"_header.json", headerJson, 0644)

	return nil
}

func traceTx(ctx *cli.Context) error {
	tx := ctx.String("tx")
	ethArchiveRemote := ctx.String("eth-archive-remote")

	rpcClient, err := rpc.Dial(ethArchiveRemote)
	if err != nil {
		log.WithError(err).Fatal("dial archive error")
	}

	var receiptClient *ethclient.Client
	if ctx.String("eth-remote") != "" {
		receiptClient, err = ethclient.Dial(ctx.String("eth-remote"))
		if err != nil {
			log.WithError(err).Fatal("dial eth error")
		}
	}

	trace, callFrame, header, err := trace(common.HexToHash(tx), rpcClient, receiptClient)

	if err != nil {
		log.WithError(err).Fatal("trace error")
	}

	if err := serialize(trace, callFrame, header, common.HexToHash(tx)); err != nil {
		log.WithError(err).Fatal("serialize error")
	}

	return nil
}
