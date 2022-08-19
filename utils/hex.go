package utils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func HexToBigInt(hexString string) (*big.Int, error) {
	value := new(big.Int)
	hexString = strings.TrimPrefix(hexString, "0x")
	_, ok := value.SetString(hexString, 16)
	if !ok {
		return value, fmt.Errorf("could not transform hex string: %s to big int", hexString)
	}

	return value, nil
}

func HexStrToInt64(hexString string) (int64, error) {
	if !strings.HasPrefix(hexString, "0x") {
		hexString = "0x" + hexString
	}
	return strconv.ParseInt(hexString, 0, 64)
}
