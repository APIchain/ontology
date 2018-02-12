package account

import (
	"os"
	"path"
	"testing"
	"fmt"
)

//run this test first
func TestClient(t *testing.T) {
	t.Log("created client start!")

	dir := "./data/"
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		t.Log("create dir ", dir, " error: ", err)
	} else {
		t.Log("create dir ", dir, " success!")
	}
	for i := 0; i < 10; i++ {
		p := path.Join(dir, fmt.Sprintf("wallet%d.txt", i))
		Create(p, []byte("123456"))
	}
}

func TestOpen(t *testing.T){
	t.Log("open client start!")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	if client == nil{
		t.Error("TestOpen should return a nonnil client!")
	}
	client = Open("./data/wallet1.txt",[]byte("12345"))
	if client != nil{
		t.Error("TestOpen should return a nonnil client! with wrong passwd")
	}
	client = Open("./data/noWallet.txt",[]byte("123456"))
	if client != nil{
		t.Error("TestOpen should return a nil client")
	}
}

func TestNewClient(t *testing.T) {
	t.Log("already tested in TestCreate!")
}

/* this should be test in cli
func TestGetClient(t *testing.T) {
	t.Log("TestGetClient start!")
	client :=GetClient()
	if client == nil{
		t.Error("TestGetClient should return a client")
	}
}
*/


func TestGetBookKeepers(t *testing.T) {
	t.Log("TestGetBookKeppers start")
	//load config.json
	pks := GetBookKeepers()
	if len(pks) != 4{
		t.Error("TestGetBookKeppers should return 4 public keys")
	}

}


func TestClientImpl_CreateAccount(t *testing.T) {
	t.Log("TestClientImpl_CreateAccount start")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	acct, err:= client.CreateAccount()
	if err != nil {
		t.Error("TestClientImpl_CreateAccount error ",err.Error())
	}
	if acct == nil {
		t.Error("TestClientImpl_CreateAccount should return a acct")
	}
}

func TestClientImpl_AddContract(t *testing.T) {
	t.Log("TestClientImpl_AddContract start")
	t.Log("this is already test in TestClientImpl_CreateAccount")

}

func TestClientImpl_ChangePassword(t *testing.T) {
	t.Log("TestClientImpl_ChangePassword start")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	ret := client.ChangePassword([]byte("123456"),[]byte("111111"))
	if !ret {
		t.Fatal("TestClientImpl_ChangePassword failed!")
	}
	client.closeDB()

	client = Open("./data/wallet0.txt",[]byte("111111"))
	if client == nil{
		t.Fatal("TestClientImpl_ChangePassword failed")
	}
	t.Log("reset password to '123456'")
	client.ChangePassword([]byte("111111"),[]byte("123456"))
}

func TestClientImpl_GetDefaultAccount(t *testing.T) {
	t.Log("TestClientImpl_GetDefaultAccount start")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	acct ,err := client.GetDefaultAccount()
	if err != nil {
		t.Fatal("TestClientImpl_GetDefaultAccount error !",err.Error())
	}
	if acct == nil{
		t.Fatal("TestClientImpl_GetDefaultAccount return nil account !")
	}
	t.Log("TestClientImpl_GetDefaultAccount end")

}

func TestClientImpl_ContainsAccount(t *testing.T) {
	t.Log("TestClientImpl_ContainsAccount start")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	for pk,_ := range client.accounts{
		acct := client.GetAccountByProgramHash(pk)
		if !client.ContainsAccount(acct.PublicKey){
			t.Errorf("TestClientImpl_ContainsAccount error! Should contains the %v",acct.PublicKey)
		}
	}
	t.Log("TestClientImpl_ContainsAccount end")
}

func TestClientImpl_CreateAccountByPrivateKey(t *testing.T) {
	t.Log("TestClientImpl_CreateAccountByPrivateKey start")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	tmpPrivKey := []byte("12345678901234567890123456789012")
	acct,err:=client.CreateAccountByPrivateKey(tmpPrivKey)

	if err!= nil{
		t.Fatal("TestClientImpl_CreateAccountByPrivateKey failed:",err.Error())
	}

	if acct == nil{
		t.Fatal("TestClientImpl_CreateAccountByPrivateKey failed: should return a account")
	}

	acctPrivkey := acct.PrivateKey
	if len(tmpPrivKey) != len(acctPrivkey){
		t.Fatal("TestClientImpl_CreateAccountByPrivateKey failed:private key not same!")
	}

	for i,b := range acctPrivkey{
		if tmpPrivKey[i] != b{
			t.Fatal("TestClientImpl_CreateAccountByPrivateKey failed:private key not same!")
		}
	}



	t.Log("TestClientImpl_CreateAccountByPrivateKey end")
}

func TestClientImpl_GetAccount(t *testing.T) {
	t.Log("TestClientImpl_GetAccount start")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	acct,_ := client.GetDefaultAccount()
	pubkey := acct.PublicKey

	tmpacct ,err := client.GetAccount(pubkey)
	if err != nil{
		t.Fatal("TestClientImpl_GetAccount failed:",err.Error())
	}
	if tmpacct == nil {
		t.Fatal("TestClientImpl_GetAccount failed: should return a account!")
	}
	if acct != tmpacct{
		t.Fatal("TestClientImpl_GetAccount failed: account not same!")
	}

	/* need a non nil pubkey???
	_, err = client.GetAccount(nil)
	if err == nil{
		t.Fatal("TestClientImpl_GetAccount failed: should return a error!")
	}
	*/

	t.Log("TestClientImpl_GetAccount end")

}

func TestClientImpl_GetAccountByProgramHash(t *testing.T) {
	t.Log("TestClientImpl_GetAccountByProgramHash start")

	client := Open("./data/wallet0.txt",[]byte("123456"))
	tmpacct,_ := client.GetDefaultAccount()
	ph := tmpacct.ProgramHash

	acct := client.GetAccountByProgramHash(ph)
	if acct != tmpacct {
		t.Error("TestClientImpl_GetAccountByProgramHash failed!")
	}

	t.Log("TestClientImpl_GetAccountByProgramHash end")
}

func TestClientImpl_LoadAccount(t *testing.T) {
	t.Log("TestClientImpl_GetAccountByProgramHash start")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	accounts := client.LoadAccount()
	if len(accounts) == 0{
		t.Error("TestClientImpl_LoadAccount failed! no accounts found!")
	}


	t.Log("TestClientImpl_GetAccountByProgramHash end")
}

func TestClientImpl_GetContract(t *testing.T) {
	t.Log("TestClientImpl_GetAccountByProgramHash start")
	client := Open("./data/wallet0.txt",[]byte("123456"))
	tmpAcct,_ := client.CreateAccount()
	ph := tmpAcct.ProgramHash

	fmt.Println(ph)

	contract := client.GetContract(ph)
	if contract == nil{
		t.Fatal("TestClientImpl_GetAccountByProgramHash failed !")
	}

	if contract.ProgramHash != ph {
		t.Error("TestClientImpl_GetAccountByProgramHash failed :ProgramHash not same!")
	}

	t.Log("TestClientImpl_GetAccountByProgramHash end")
}

func TestClientImpl_LoadContracts(t *testing.T) {
	t.Log("TestClientImpl_LoadContracts start")

	client := Open("./data/wallet0.txt",[]byte("123456"))

	contracts := client.LoadContracts()
	if len(contracts) == 0 {
		t.Error("TestClientImpl_LoadContracts failed :no contracts loaded")
	}

	t.Log("TestClientImpl_LoadContracts end")
}