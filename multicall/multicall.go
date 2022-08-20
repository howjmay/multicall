package multicall

import (
	"encoding/hex"

	"github.com/howjmay/multicall/ethrpc"
)

type Multicall interface {
	CallRaw(calls ViewCalls, block string) (*Result, error)
	Call(calls ViewCalls, block string) (*Result, error)
	Contract() string
}

type multicall struct {
	eth    ethrpc.ETHInterface
	config *Config
}

func New(eth ethrpc.ETHInterface, opts ...Option) (Multicall, error) {
	config := &Config{
		MulticallAddress: MainnetAddress,
		Gas:              "0x400000000",
	}

	for _, opt := range opts {
		opt(config)
	}

	return &multicall{
		eth:    eth,
		config: config,
	}, nil
}

type CallResult struct {
	Success bool
	Raw     []byte
	Decoded []interface{}
}

type Result struct {
	BlockNumber uint64
	Calls       map[string]CallResult
}

const AggregateMethod = "0x17352e13"

func (mc multicall) CallRaw(calls ViewCalls, block string) (*Result, error) {
	resultRaw, err := mc.sendRequest(calls, block)
	if err != nil {
		return nil, err
	}
	return calls.decodeRaw(resultRaw)
}

func (mc multicall) Call(calls ViewCalls, block string) (*Result, error) {
	resultRaw, err := mc.sendRequest(calls, block)
	if err != nil {
		return nil, err
	}
	return calls.decode(resultRaw)
}

func (mc multicall) sendRequest(calls ViewCalls, block string) (string, error) {
	payloadArgs, err := calls.callData()
	if err != nil {
		return "", err
	}
	payload := make(map[string]string)
	payload["to"] = mc.config.MulticallAddress
	payload["data"] = AggregateMethod + hex.EncodeToString(payloadArgs)
	payload["gas"] = mc.config.Gas
	var resultRaw string
	err = mc.eth.SendRequest(&resultRaw, ethrpc.ETH_Call, payload, block)
	return resultRaw, err
}

func (mc multicall) Contract() string {
	return mc.config.MulticallAddress
}
