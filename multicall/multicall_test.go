package multicall_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/howjmay/multicall/ethrpc"
	"github.com/howjmay/multicall/ethrpc/provider/httprpc"
	"github.com/howjmay/multicall/multicall"
	"github.com/stretchr/testify/require"
)

func TestExampleViwCall(t *testing.T) {
	eth, err := getETH("https://mainnet.infura.io/v3/17ed7fe26d014e5b9be7dfff5368c69d")
	require.NoError(t, err)
	vc := multicall.NewViewCall(
		"key.1",
		"0x5d3a536E4D6DbD6114cc1Ead35777bAB948E3643",
		"totalReserves()(uint256)",
		[]interface{}{},
	)
	vcs := multicall.ViewCalls{vc}
	mc, _ := multicall.New(eth)
	block := "latest"
	res, err := mc.Call(vcs, block)
	require.NoError(t, err)

	resJson, _ := json.Marshal(res)
	fmt.Println(string(resJson))
	fmt.Println(res)
	fmt.Println(err)

}

func getETH(url string) (ethrpc.ETHInterface, error) {
	provider, err := httprpc.New(url)
	if err != nil {
		return nil, err
	}
	provider.SetHTTPTimeout(5 * time.Second)
	return ethrpc.New(provider)
}
