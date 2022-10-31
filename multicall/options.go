package multicall

import "fmt"

type Option func(*Config)

type Config struct {
	MulticallAddress string
	Gas              string
}

const (
	// MainnetMulticall : Multicall contract address on mainnet
	MainnetAddress = "0xeefba1e63905ef1d7acba5a8513c70307c1ce441"
	// GoerliMulticall : Multicall contract address on Goerli
	GoerliAddress = "0x77dca2c955b15e9de4dbbcf1246b4b85b651e50e"
)

func ContractAddress(address string) Option {
	return func(c *Config) {
		c.MulticallAddress = address
	}
}

func SetGas(gas uint64) Option {
	return func(c *Config) {
		c.Gas = fmt.Sprintf("0x%x", gas)
	}
}

func SetGasHex(gas string) Option {
	return func(c *Config) {
		c.Gas = gas
	}
}
