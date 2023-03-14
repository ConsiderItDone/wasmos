package polywrapvm

import (
	"fmt"
	wasmvm "github.com/CosmWasm/wasmvm"
	"github.com/polywrap/go-client/msgpack"
)

type CosmosPlugin struct {
	store wasmvm.KVStore
}

type DbWriteArgType struct {
	Key   []byte
	Value []byte
}

func NewCosmosPlugin() *CosmosPlugin {
	return &CosmosPlugin{}
}

func (cp *CosmosPlugin) DbWrite(args DbWriteArgType) int32 {
	cp.store.Set(args.Key, args.Value)
	return 1
}

func (cp *CosmosPlugin) SetStore(store wasmvm.KVStore) {
	cp.store = store
}

func (cp *CosmosPlugin) EncodeArgs(method string, args []byte) (any, error) {
	switch method {
	case "DbWrite":
		return msgpack.Decode[DbWriteArgType](args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}
