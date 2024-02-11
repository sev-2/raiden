package state

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
)

type RpcStructField struct {
	Name string
	Type string
	Tag  string
}

type RpcFunction struct {
	Name string
	Body string
}

type RpcTypeDef struct {
	Name        string
	Type        string
	ArrType     string
	Fields      []RpcStructField
	RpcFunction []RpcFunction
}

const (
	RpcTypeDevArray  = "array"
	RpcTypeDevStruct = "struct"
)

func ToRpc(rpcState []RpcState, appRpc []raiden.Rpc) (listRpc []supabase.Function, err error) {
	return
}
