package multicall

import (
	"fmt"
	"math/big"
)

type BigIntJSONString big.Int

func (bi BigIntJSONString) MarshalJSON() ([]byte, error) {
	backToInt := big.Int(bi)
	return []byte(fmt.Sprintf(`"%s"`, backToInt.String())), nil
}

func (bi BigIntJSONString) String() string {
	gobi := big.Int(bi)
	return (&gobi).String()
}

func (bi BigIntJSONString) ToBigInt() *big.Int {
	gobi := big.Int(bi)
	return &gobi
}
