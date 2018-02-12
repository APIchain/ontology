package account

import (
	"github.com/Ontology/crypto"

	"testing"
	"github.com/Ontology/common/log"
)

func init()  {
	log.Init(log.Path, log.Stdout)
	crypto.SetAlg("P256R1")
}


func TestNewAccount(t *testing.T) {
	t.Log("NewAccount start!")
	acct,err:=NewAccount()
	if err != nil{
		t.Error("NewAccount error:",err.Error())
	}
	if acct == nil{
		t.Error("NewAccount nil!")
	}
	t.Logf("acct is %v",acct)
}

func TestNewAccountWithPrivatekey(t *testing.T){
	t.Log("TestNewAccountWithPrivatekey start!")
	wrongKey1 := []byte("")

	_,err := NewAccountWithPrivatekey(wrongKey1)
	if err == nil{
		t.Error("NewAccountWithPrivatekey should return Error when input is " ,string(wrongKey1))
	}

	key32 := []byte("12345678901234567890123456789012")
	_,err = NewAccountWithPrivatekey(key32)
	if err != nil{
		t.Error("NewAccountWithPrivatekey return error:",err.Error())
	}

	key96 := []byte("123456789012345678901234567890121234567890123456789012345678901212345678901234567890123456789012")
	_,err = NewAccountWithPrivatekey(key96)
	if err != nil{
		t.Error("NewAccountWithPrivatekey return error:",err.Error())
	}

	key104 := []byte("12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234")
	_,err = NewAccountWithPrivatekey(key104)
	if err != nil{
		t.Error("NewAccountWithPrivatekey return error:",err.Error())
	}
	t.Log("Test succeed!")
}