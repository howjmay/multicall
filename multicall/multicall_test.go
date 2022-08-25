package multicall_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/howjmay/multicall/multicall"
	"github.com/stretchr/testify/require"
)

func TestExampleViewCall(t *testing.T) {
	eth, err := multicall.GetETH("https://rpc.ankr.com/eth")
	require.NoError(t, err)
	vc := multicall.NewViewCall(
		"SHIB-symbol",
		"0x95aD61b0a150d79219dCF64E1E6Cc01f0B64C4cE", // shib contract address
		"symbol()(string)",
		[]interface{}{},
	)
	vcs := multicall.ViewCalls{vc}
	mc, _ := multicall.New(eth)
	block := "latest"
	res, err := mc.Call(vcs, block)
	require.NoError(t, err)
	require.True(t, res.Calls["SHIB-symbol"].Success)
	symbolRes := res.Calls["SHIB-symbol"].Decoded[0].(string)
	require.Equal(t, "SHIB", symbolRes)
	require.Len(t, res.Calls, 1)

	resJson, err := json.Marshal(res)
	require.NoError(t, err)
	fmt.Println(string(resJson))
	fmt.Println(res)
}
