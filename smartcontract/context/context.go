package context

import (
	"github.com/Ontology/common"
	vmtypes "github.com/Ontology/vm/types"
)

type ContextRef interface {
	PushContext(context *Context)
	CurrentContext() *Context
	CallingContext() *Context
	EntryContext() *Context
	PopContext()
	CheckWitness(address common.Address) bool
	Execute() error
}


type Context struct {
	ContractAddress common.Address
	Code vmtypes.VmCode
}
