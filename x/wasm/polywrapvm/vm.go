package polywrapvm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CosmWasm/wasmvm/types"
	"github.com/polywrap/go-client/wasm"
	polywrapClient "github.com/polywrap/go-client/wasm/client"
	"github.com/polywrap/go-client/wasm/uri"
	"os"
	"path/filepath"

	wasmvm "github.com/CosmWasm/wasmvm"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const wasmDir = "wasm"

type VM struct {
	dataDir string
	client  *polywrapClient.Client
}

type ArgsInstantiate struct {
	name string
}
type InstantiateResult struct {
	Result string
}

func NewVM(dataDir string) (*VM, error) {
	wasmPath := filepath.Join(dataDir, wasmDir)
	err := os.MkdirAll(wasmPath, 0755)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to create wasm directory")
	}

	client := polywrapClient.New(&polywrapClient.ClientConfig{
		Resolver: wasm.NewFsResolver(),
	})

	return &VM{
		dataDir: dataDir,
		client:  client,
	}, nil

}
func (vm *VM) Create(code wasmvm.WasmCode) (wasmvm.Checksum, error) {
	if code == nil {
		return nil, errors.New("wasm code couldn't be nil")
	}
	checksum := sha256.Sum256(code)
	encodedChecksum := hex.EncodeToString(checksum[:])

	path := filepath.Join(vm.dataDir, wasmDir, encodedChecksum)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to create wasm directory")
	}

	// write wrap file
	err = os.WriteFile(vm.getWasmFilePath(checksum[:]), code, 0755)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to write wasm file")
	}

	// write fake manifest file
	err = os.WriteFile(vm.getManifestFilePath(checksum[:]), []byte(""), 0755)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to write wasm file")
	}

	return checksum[:], nil
}

func (vm *VM) AnalyzeCode(_ wasmvm.Checksum) (*types.AnalysisReport, error) {
	return &types.AnalysisReport{
		HasIBCEntryPoints:    false,
		RequiredFeatures:     "",
		RequiredCapabilities: "",
	}, nil
}

func (vm *VM) GetCode(checksum wasmvm.Checksum) (wasmvm.WasmCode, error) {
	wrapper, err := os.ReadFile(vm.getWasmFilePath(checksum))
	if err != nil {
		return nil, err
	}

	return wrapper, nil
}

func (vm *VM) Instantiate(checksum wasmvm.Checksum, env types.Env, info types.MessageInfo, initMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost types.UFraction) (*types.Response, uint64, error) {
	wrapperPath := "wrap://fs/" + vm.getWasmFileDir(checksum)
	wrapperUri, err := uri.New(wrapperPath)
	if err != nil {
		return nil, 0, err
	}

	//args := ArgsInstantiate{
	//	name: "zzz",
	//}
	gasUsed := uint64(100) //10847  99149

	var args map[string]interface{}
	err = json.Unmarshal(initMsg, &args)
	if err != nil {
		return nil, 0, sdkerrors.Wrap(err, "unable to unmarshal init message")
	}

	res, err := polywrapClient.Invoke[map[string]interface{}, InstantiateResult, []byte](vm.client, *wrapperUri, "instantiate", args, nil)
	if err != nil {
		return nil, gasUsed, err
	}

	return &types.Response{
		Messages:   nil,
		Data:       []byte(res.Result),
		Attributes: nil,
		Events:     nil,
	}, gasUsed, nil

	//gasForDeserialization := deserCost.Mul(uint64(len(data))).Floor()
	//if gasLimit < gasForDeserialization+gasUsed {
	//	return nil, gasUsed, fmt.Errorf("Insufficient gas left to deserialize contract execution result (%d bytes)", len(data))
	//}
	//gasUsed += gasForDeserialization

	//var result types.ContractResult
	//err = json.Unmarshal(data, &result)
	//if err != nil {
	//	return nil, gasUsed, err
	//}
	//if result.Err != "" {
	//	return nil, gasUsed, fmt.Errorf("%s", result.Err)
	//}
	//return result.Ok, gasUsed, nil
}

func (vm *VM) Execute(checksum wasmvm.Checksum, env types.Env, info types.MessageInfo, executeMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost types.UFraction) (*types.Response, uint64, error) {
	envBin, err := json.Marshal(env)
	if err != nil {
		return nil, 0, err
	}
	infoBin, err := json.Marshal(info)
	if err != nil {
		return nil, 0, err
	}
	data, gasUsed, err := api.Execute(vm.cache, checksum, envBin, infoBin, executeMsg, &gasMeter, store, &goapi, &querier, gasLimit, vm.printDebug)
	if err != nil {
		return nil, gasUsed, err
	}

	gasForDeserialization := deserCost.Mul(uint64(len(data))).Floor()
	if gasLimit < gasForDeserialization+gasUsed {
		return nil, gasUsed, fmt.Errorf("Insufficient gas left to deserialize contract execution result (%d bytes)", len(data))
	}

	gasUsed += gasForDeserialization
	var result types.ContractResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, gasUsed, err
	}
	if result.Err != "" {
		return nil, gasUsed, fmt.Errorf("%s", result.Err)
	}
	return result.Ok, gasUsed, nil
}

func (vm *VM) getWasmFilePath(checksum wasmvm.Checksum) string {
	return filepath.Join(vm.getWasmFileDir(checksum), "wrap.wasm")
}

func (vm *VM) getManifestFilePath(checksum wasmvm.Checksum) string {
	return filepath.Join(vm.getWasmFileDir(checksum), "wrap.info")
}

func (vm *VM) getWasmFileDir(checksum wasmvm.Checksum) string {
	return filepath.Join(vm.dataDir, wasmDir, hex.EncodeToString(checksum))
}
