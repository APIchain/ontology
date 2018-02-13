package neovm

import (
	"math/big"
	"testing"
	"fmt"
)

const(
	INT_MAX = int(^uint(0) >> 1)
	INT_MIN = ^INT_MAX
)

func TestCommon(t *testing.T) {
	i := ToBigInt(big.NewInt(1))
	t.Log("i", i)

	fmt.Println(ToArrayReverse([]byte{1, 2, 3}))
}

func ToArrayReverse(arr []byte) []byte {
	l := len(arr)
	x := make([]byte, 0)
	for i := l - 1; i >= 0; i-- {
		x = append(x, arr[i])
	}
	return x
}

func TestBigIntMultiComp(t *testing.T) {

	b1,result := new(big.Int).SetString("1000000000000000000000000000000000",10)
	if !result {
		t.Fatal("error")
	}
	b2,result := new(big.Int).SetString("1000000000000000000000000000000000",10)
	res := BigIntMultiComp(b1,b2,NUMEQUAL)
	if !res {
		t.Error("BigIntMultiComp failed")
	}
	b2 ,_ = new(big.Int).SetString("1000000000000000000000000000000001",10)

	res = BigIntMultiComp(b1,b2 ,NUMNOTEQUAL)
	if !res {
		t.Error("BigIntMultiComp failed")
	}

}