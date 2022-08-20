package ethrpc

// json rpc methods
const (
	// parity
	Parity_Subscribe           = "parity_subscribe"
	Parity_PendingTransactions = "parity_pendingTransactions"

	// geth
	GETH_TxPoolContent = "txpool_content"

	// net
	Net_PeerCount = "net_peerCount"

	// web3
	WEB3_ClientVersion = "web3_clientVersion"

	// eth
	ETH_BlockNumber                      = "eth_blockNumber"
	ETH_Call                             = "eth_call"
	ETH_GetBalance                       = "eth_getBalance"
	ETH_GetBlockByNumber                 = "eth_getBlockByNumber"
	ETH_GetBlockTransactionCountByNumber = "eth_getBlockTransactionCountByNumber"
	ETH_GetCode                          = "eth_getCode"
	ETH_GetFilterChanges                 = "eth_getFilterChanges"
	ETH_GetTransactionByHash             = "eth_getTransactionByHash"
	ETH_GetTransactionReceipt            = "eth_getTransactionReceipt"
	ETH_GetUncleByBlockHashAndIndex      = "eth_getUncleByBlockHashAndIndex"
	ETH_GetUncleByBlockNumberAndIndex    = "eth_getUncleByBlockNumberAndIndex"
	ETH_PendingTransactionFilter         = "eth_newPendingTransactionFilter"
	ETH_Subscribe                        = "eth_subscribe"

	// trace
	Trace_Block                   = "trace_block"
	Trace_ReplayBlockTransactions = "trace_replayBlockTransactions"

	// eth pubsub
	ETH_NewHeads               = "newHeads"
	ETH_NewPendingTransactions = "newPendingTransactions"

	// client
	Client_GETH   = "geth"
	Client_Parity = "parity"
)

// ERC20 signatures
const (
	// functions
	NameFunction         = "0x06fdde03"
	ApproveFunction      = "0x095ea7b3" // mandatory
	TotalSupplyFunction  = "0x18160ddd" // mandatory
	TransferFromFunction = "0x23b872dd" // mandatory
	DecimalsFunction     = "0x313ce567"
	IssueTokensFunction  = "0x475a9fa9"
	BalanceOfFunction    = "0x70a08231" // mandatory
	SymbolFunction       = "0x95d89b41"
	TransferFunction     = "0xa9059cbb" // mandatory
	AllowanceFunction    = "0xdd62ed3e" // mandatory

	// events
	TransferEvent = "0xddf252ad" // mandatory
	ApprovalEvent = "0x8c5be1e5" // mandatory
)

const (
	// DefaultCallGas is the default gas to use for eth_calls
	DefaultCallGas = "0xffffff"
)
