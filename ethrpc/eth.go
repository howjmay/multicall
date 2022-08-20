package ethrpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/howjmay/multicall/types"
	"github.com/howjmay/multicall/utils"

	"github.com/howjmay/multicall/errors"
	"github.com/howjmay/multicall/ethrpc/provider"
	"github.com/howjmay/multicall/ethrpc/provider/httprpc"
	"github.com/howjmay/multicall/ethrpc/provider/wsrpc"
)

// ETH server interaction
type ETH struct {
	rpc    provider.Interface
	client string
}

// Start connects to parity and starts listening for notifications
func (e *ETH) Start() error {
	// connect and start read/write pumps
	err := e.rpc.Start()
	if err != nil {
		return err
	}

	// find out if this is geth or parity
	c, err := e.GetClient()
	e.client = c
	return err
}

// Stop closes the connection
func (e *ETH) Stop() {
	e.rpc.Stop()
}

// GetLatestBlock with or without full transactions array
func (e *ETH) GetLatestBlock() (b types.Block, err error) {
	err = e.SendRequest(&b, ETH_GetBlockByNumber, "latest", true)
	return
}

// GetBlockByNumber gets specified block with full transaction array
func (e *ETH) GetBlockByNumber(number string) (b types.Block, err error) {
	err = e.SendRequest(&b, ETH_GetBlockByNumber, number, true)
	return
}

// GetBlockTransactionCountByNumber https://wiki.parity.io/JSONRPC-eth-module.html#eth_getblocktransactioncountbynumber
func (e *ETH) GetBlockTransactionCountByNumber(number string) (count string, err error) {
	err = e.SendRequest(&count, ETH_GetBlockTransactionCountByNumber, number)
	return
}

// GetUncleByBlockHashAndIndex retrieves the index-nth uncle of the
// block with the hash blockHash
func (e *ETH) GetUncleByBlockHashAndIndex(hash string, index string) (b types.Block, err error) {
	err = e.SendRequest(&b, ETH_GetUncleByBlockHashAndIndex, hash, index)
	return
}

// GetUncleByBlockNumberAndIndex retrieves the index-nth uncle of the
// block with the number blockNumber
func (e *ETH) GetUncleByBlockNumberAndIndex(blockNumber string, index string) (b types.Block, err error) {
	err = e.SendRequest(&b, ETH_GetUncleByBlockNumberAndIndex, blockNumber, index)
	return
}

// GetPeerCount gets current peer count
func (e *ETH) GetPeerCount() (peers int64, err error) {
	var ps string
	err = e.SendRequest(&ps, Net_PeerCount)
	if err != nil {
		return
	}
	peers, err = strconv.ParseInt(ps, 0, 64)
	return
}

// GetVersion gets current eth client version string
func (e *ETH) GetVersion() (ver string, err error) {
	err = e.SendRequest(&ver, WEB3_ClientVersion)
	return ver, err
}

// GetClient gets current eth client version string
func (e *ETH) GetClient() (string, error) {
	var client string
	if e.client != "" {
		return e.client, nil
	}
	c, err := e.GetVersion()
	if err == nil {
		c = strings.ToLower(c)
		if strings.HasPrefix(c, Client_GETH) {
			client = Client_GETH
		} else if strings.HasPrefix(c, Client_Parity) {
			client = Client_Parity
		}
	}
	return client, err
}

// SetPendingTransactionsFilter sets pending transaction filter for ETHGetFilterChanges
func (e *ETH) SetPendingTransactionsFilter() (id string, err error) {
	err = e.SendRequest(&id, ETH_PendingTransactionFilter)
	return
}

// GetFilterChanges gets filtered entities, since last poll or set filter
func (e *ETH) GetFilterChanges(id string) (t []interface{}, err error) {
	err = e.SendRequest(&t, ETH_GetFilterChanges, id)
	return
}

// GetPendingFilterChanges gets all pending transactions, filtered, since last poll or set filter
func (e *ETH) GetPendingFilterChanges(id string) (t []string, err error) {
	err = e.SendRequest(&t, ETH_GetFilterChanges, id)
	return
}

// GetPendingTransactions gets full array of pending transactions
func (e *ETH) GetPendingTransactions() ([]types.Transaction, error) {
	var txs []types.Transaction
	var err error
	if e.client == Client_Parity {
		err = e.SendRequest(&txs, Parity_PendingTransactions)
	} else if e.client == Client_GETH {
		var pool types.GethTxPool
		err = e.SendRequest(&pool, GETH_TxPoolContent)

		for typ := range pool {
			for addr := range pool[typ] {
				for _, tx := range pool[typ][addr] {
					if tx.BlockHash == "0x0000000000000000000000000000000000000000000000000000000000000000" {
						tx.BlockHash = ""
					}
					txs = append(txs, tx)
				}
			}
		}
	}
	return txs, err
}

// GetTransactionByHash gets a transaction by transaction hash
func (e *ETH) GetTransactionByHash(hash string) (types.Transaction, error) {
	var t types.Transaction
	err := e.SendRequest(&t, ETH_GetTransactionByHash, hash)
	// geth correction
	if t.BlockNumber == "" && t.BlockHash == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		t.BlockHash = ""
	}
	return t, err
}

// GetTransactionReceipt gets the transaction receipt ofr a specific transaction hash
func (e *ETH) GetTransactionReceipt(hash string) (r types.Receipt, err error) {
	err = e.SendRequest(&r, ETH_GetTransactionReceipt, hash)
	return
}

// GetRawBalanceAtBlock returns the balance of an address at a given blockNumber as a hex string
func (e *ETH) GetRawBalanceAtBlock(address, blockNumber string) (string, error) {
	var result string
	err := e.SendRequest(&result, ETH_GetBalance, address, blockNumber)
	if err != nil {
		return "", err
	}
	if result == "0x" || result == "" {
		return "", errors.Empty
	}
	return result, nil
}

// GetBalanceAtBlock returns the balance of an address at a given blockNumber as a big.Int
func (e *ETH) GetBalanceAtBlock(address, blockNumber string) (*big.Int, error) {
	rawBalance, err := e.GetRawBalanceAtBlock(address, blockNumber)
	if err != nil {
		return nil, err
	}
	return utils.HexToBigInt(rawBalance)
}

// GetRawTokenBalanceAtBlock returns the token balance of an address at a given blockNumber as a hex string
func (e *ETH) GetRawTokenBalanceAtBlock(address, token, blockNumber string) (string, error) {
	var result string
	payload := make(map[string]string)
	payload["to"] = token
	payload["data"] =
		BalanceOfFunction +
			strings.Repeat("0", 32-len(BalanceOfFunction)+2) +
			strings.Replace(address, "0x", "", 1)
	err := e.SendRequest(&result, ETH_Call, payload, blockNumber)
	if err != nil {
		return "", err
	}
	if result == "0x" || result == "" {
		return "", errors.Empty
	}
	return result, nil
}

// GetTokenBalanceAtBlock returns the token balance of an address at a given blockNumber as a big.Int
func (e *ETH) GetTokenBalanceAtBlock(address, token, blockNumber string) (*big.Int, error) {
	rawBalance, err := e.GetRawTokenBalanceAtBlock(address, token, blockNumber)
	if err != nil {
		return nil, err
	}
	return utils.HexToBigInt(rawBalance)
}

// GetBlockNumber returns the number of most recent block.
func (e *ETH) GetBlockNumber() (int64, error) {
	var n string
	err := e.SendRequest(&n, ETH_BlockNumber)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(n, 0, 64)
}

// GetContractName calls a contract's name function
func (e *ETH) GetContractName(address string) (string, error) {
	return e.CallContractFunction(NameFunction, address, DefaultCallGas)
}

// GetContractSymbol calls a contract's name function
func (e *ETH) GetContractSymbol(address string) (string, error) {
	return e.CallContractFunction(SymbolFunction, address, DefaultCallGas)
}

// GetContractTotalSupply calls a contract's totalSupply function
func (e *ETH) GetContractTotalSupply(address string) (*big.Int, error) {
	return e.CallContractFunctionBigInt(TotalSupplyFunction, address)
}

// GetERC20Decimals calls a contract's decimal function
func (e *ETH) GetERC20Decimals(address string) (uint8, error) {
	d, err := e.CallContractFunctionInt64(DecimalsFunction, address)
	if err != nil {
		switch err.(type) {
		case *strconv.NumError:
			return 0, errors.InvalidUInt8
		default:
			return 0, err
		}

	}
	if d < 0 || d > 255 {
		return 0, errors.InvalidUInt8
	}
	return uint8(d), nil
}

// GetCode returns the bytecode of a contract
func (e *ETH) GetCode(a string) ([]byte, error) {
	var s string
	err := e.SendRequest(&s, ETH_GetCode, a, "latest")
	if err != nil {
		return nil, err
	}

	s = strings.TrimPrefix(s, "0x")
	return hex.DecodeString(s)
}

// Traces
func (e *ETH) TraceBlock(blockNumber string) ([]types.Trace, error) {
	var traces []types.Trace
	err := e.SendRequest(&traces, Trace_Block, blockNumber)
	return traces, err
}

func (e *ETH) TraceReplayBlockTransactions(blockNumber string, traceTypes ...string) ([]types.TransactionReplay, error) {
	var replays []types.TransactionReplay
	err := e.SendRequest(&replays, Trace_ReplayBlockTransactions, blockNumber, traceTypes)
	return replays, err
}

// NewHeadsSubscription eth_subscribe to newHeads
func (e *ETH) NewHeadsSubscription() (r chan *types.BlockHeader, err error) {
	r = make(chan *types.BlockHeader, 100)
	j := make(chan *json.RawMessage, 100)

	go func(res chan<- *types.BlockHeader, notif <-chan *json.RawMessage) {
		for notification := range notif {
			var blockHead types.BlockHeader
			err := json.Unmarshal(*notification, &blockHead)
			if err != nil {
				log.Fatal("unmarshal notification", err)
			}
			res <- &blockHead
		}
		close(res)
	}(r, j)

	err = e.Subscribe(j, ETH_Subscribe, ETH_NewHeads)
	return
}

// NewPendingTransactionsSubscription eth_subscribe to newPendingTransactions
func (e *ETH) NewPendingTransactionsSubscription() (r chan *string, err error) {
	r = make(chan *string, 10000)
	j := make(chan *json.RawMessage, 10000)

	go func(res chan<- *string, notif <-chan *json.RawMessage) {
		for notification := range notif {
			var tx string
			err := json.Unmarshal(*notification, &tx)
			if err != nil {
				log.Fatal("unmarshal notification", err)
			}
			res <- &tx
		}
		close(res)
	}(r, j)

	err = e.Subscribe(j, ETH_Subscribe, ETH_NewPendingTransactions)
	return
}

// NewBlockNumberSubscription parity_subscribe to eth_blockNumber
func (e *ETH) NewBlockNumberSubscription() (r chan *int64, err error) {
	r = make(chan *int64, 10000)
	j := make(chan *json.RawMessage, 10000)

	go func(res chan<- *int64, notif <-chan *json.RawMessage) {
		for notification := range notif {
			var bn string
			err := json.Unmarshal(*notification, &bn)
			if err != nil {
				log.Fatal("unmarshal notification", err)
			}
			n, err := strconv.ParseInt(bn, 0, 64)
			if err != nil {
				log.Fatal("convert notification", err)
			}
			res <- &n
		}
		close(res)
	}(r, j)

	err = e.Subscribe(j, Parity_Subscribe, ETH_BlockNumber, []string{})
	return
}

// CallContractFunctionInt64 calls a contract's function and returns a decoded int64
func (e *ETH) CallContractFunctionInt64(function string, address string) (int64, error) {
	ba, err := e.CallContractFunction(function, address, DefaultCallGas)
	if err != nil {
		return 0, err
	}

	return utils.HexStrToInt64(string(ba))
}

// CallContractFunctionBigInt calls a contract's function and returns a decoded int64
func (e *ETH) CallContractFunctionBigInt(function string, address string) (*big.Int, error) {
	ba, err := e.CallContractFunction(function, address, DefaultCallGas)
	if err != nil {
		return nil, err
	}

	return utils.HexToBigInt(string(ba))
}

// CallContractFunction calls a contract's function and returns the result as string
func (e *ETH) CallContractFunction(function string, address string, gas string) (string, error) {
	var s string
	obj := make(map[string]string)
	obj["to"] = address
	obj["data"] = function
	obj["gas"] = gas
	err := e.SendRequest(&s, ETH_Call, obj, "latest")
	if s == "0x" {
		return "", errors.Empty
	}
	return s, err
}

// SendRequest to server
func (e *ETH) SendRequest(result interface{}, method string, params ...interface{}) error {
	return e.rpc.Call(&result, method, params...)
}

// SendRequestRaw to server
func (e *ETH) SendRequestRaw(method string, params ...interface{}) ([]byte, error) {
	return e.rpc.CallRaw(method, params...)
}

// Subscribe to topic
func (e *ETH) Subscribe(receiver chan *json.RawMessage, method string, event string, params ...interface{}) error {
	return e.rpc.Subscribe(receiver, method, event, params...)
}

// New create a new ethereum server json rpc interface
func New(provider provider.Interface) (*ETH, error) {
	return &ETH{
			rpc: provider,
		},
		nil
}

// NewWithDefaults selects the proper provider based on protocol
func NewWithDefaults(url string) (*ETH, error) {
	switch {
	case strings.HasPrefix(url, "http"):
		p, err := httprpc.New(url)
		if err != nil {
			return nil, err
		}
		e, err := New(p)
		if err != nil {
			return nil, err
		}

		return e, nil
	case strings.HasPrefix(url, "ws"):
		p, err := wsrpc.New(url, true)
		if err != nil {
			return nil, err
		}
		e, err := New(p)
		if err != nil {
			return nil, err
		}

		return e, e.Start()
	}

	return nil, fmt.Errorf("protocol not recognized, use http(s) or ws(s)")
}
