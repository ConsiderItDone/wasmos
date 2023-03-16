package polywrapvm

import (
	"fmt"
	wasmvm "github.com/CosmWasm/wasmvm"
	"github.com/polywrap/go-client/msgpack"
)

type CosmosPlugin struct {
	store wasmvm.KVStore
}

type DbSetArgType struct {
	Key   []byte
	Value []byte
}
type DbGetArgType struct {
	Key []byte
}

func NewCosmosPlugin() *CosmosPlugin {
	return &CosmosPlugin{}
}

func (cp *CosmosPlugin) DbSet(args DbSetArgType) bool {
	cp.store.Set(args.Key, args.Value)
	return true
}
func (cp *CosmosPlugin) DbGet(args DbGetArgType) []byte {
	return cp.store.Get(args.Key)
}
func (cp *CosmosPlugin) DbHas(args DbGetArgType) bool {
	return len(cp.store.Get(args.Key)) > 0
}
func (cp *CosmosPlugin) DbRemove(args DbGetArgType) bool {
	cp.store.Delete(args.Key)
	return true
}

func (cp *CosmosPlugin) SetStore(store wasmvm.KVStore) {
	cp.store = store
}

func (cp *CosmosPlugin) EncodeArgs(method string, args []byte) (any, error) {
	switch method {
	case "DbSet":
		return msgpack.Decode[DbSetArgType](args)
	case "DbGet", "DbHas", "DbRemove":
		return msgpack.Decode[DbGetArgType](args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}
