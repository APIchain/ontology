package neovm

import (
	"testing"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common"
	"fmt"
)

func init()  {
	log.Init(log.Path, log.Stdout)

}

const(
	CODE = "52c56b6153c56c766b00527ac46c766b00c35161682b53797374656d2e457865637574696f6e456e67696e652e47657443616c6c696e6753637269707448617368c46c766b00c35261682d53797374656d2e457865637574696f6e456e67696e652e476574457865637574696e6753637269707448617368c46c766b00c36c766b51527ac46203006c766b51c3616c7566"
	CODE2 = "52c56b6153c56c766b00527ac46c766b00c30061682953797374656d2e457865637574696f6e456e67696e652e476574456e74727953637269707448617368c46c766b00c35161682b53797374656d2e457865637574696f6e456e67696e652e47657443616c6c696e6753637269707448617368c46c766b00c35261682d53797374656d2e457865637574696f6e456e67696e652e476574457865637574696e6753637269707448617368c46c766b00c36c766b51527ac46203006c766b51c3616c7566"
	CODE3 = "51c56b616167271a3444931a2081fe0544cf9309a9542b9c67fe6c766b00527ac46203006c766b00c3616c7566"
	)

func TestNewExecutionEngine(t *testing.T) {
	engine := NewExecutionEngine(nil,nil,nil,nil)

	if engine == nil{
		t.Error("TestNewExecutionEngine failed")
	}

}


func TestExecutionEngine_Call(t *testing.T) {
	caller := common.Uint160{}
	engine := NewExecutionEngine(nil,nil,nil,nil)
	//engine.Call(caller,[]byte(CODE),nil)

	_, err := engine.Call(caller,[]byte(CODE),nil)
	if err != nil{
		t.Error("TestExecutionEngine_Call failed:",err.Error())
	}

/*	b := engine.GetExecuteResult()
	fmt.Println(b)
	fmt.Println(PopByteArray(engine))*/
/*	badcode := []byte("abcd121313")
	res, err = engine.Call(caller,badcode,nil)
	if err == nil{
		t.Error("TestExecutionEngine_Call failed bad code")
	}*/
}

func TestExecutionEngine_CurrentContext(t *testing.T) {
	//caller := common.Uint160{}
	engine := NewExecutionEngine(nil,nil,nil,nil)
	_ ,err:= engine.CurrentContext()
	if err == nil{
		t.Error("TestExecutionEngine_CurrentContext failed:should return an error")
	}
	engine.LoadCode([]byte(CODE),false)
	ctx,err := engine.CurrentContext()
	if err != nil{
		t.Error("TestExecutionEngine_CurrentContext failed:can't get ctx")
	}
	fmt.Println(ctx)
}

func TestExecutionEngine_AddBreakPoint(t *testing.T) {
	caller := common.Uint160{}
	engine := NewExecutionEngine(nil,nil,nil,nil)



	_, err := engine.Call(caller,[]byte(CODE),nil)
	if err != nil{
		t.Error("TestExecutionEngine_Call failed:",err.Error())
	}

	engine.AddBreakPoint(53)
/*	ctx ,_ := engine.CurrentContext()
	fmt.Println(ctx.GetInstructionPointer())*/
	engine.StepOver()
}
