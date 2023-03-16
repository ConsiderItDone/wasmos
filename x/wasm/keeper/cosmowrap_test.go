package keeper

import (
	"bytes"
	"fmt"
	"github.com/ConsiderItDone/cosmowrap/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCosmowrapInit(t *testing.T) {
	ctx, keepers := CreateTestInput(t, false, AvailableCapabilities)

	deposit := sdk.NewCoins(sdk.NewInt64Coin("denom", 100000))
	creator := sdk.AccAddress(bytes.Repeat([]byte{1}, address.Len))
	keepers.Faucet.Fund(ctx, creator, deposit...)
	example := StoreHelloWorldExampleContract(t, ctx, keepers)

	initValue := "Ramil"
	initMsgBz := HelloWorldInitMsg{
		name: initValue,
	}.GetBytes(t)

	gasBefore := ctx.GasMeter().GasConsumed()

	em := sdk.NewEventManager()
	// create with no balance is also legal
	gotContractAddr, resp, err := keepers.ContractKeeper.Instantiate(ctx.WithEventManager(em), example.CodeID, creator, nil, initMsgBz, "demo contract 1", nil)
	require.NoError(t, err)
	require.Equal(t, "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr", gotContractAddr.String())

	require.Equal(t, fmt.Sprintf("Hello %s. CosmoWrap Initialized!", initValue), string(resp))
	t.Logf("Initialize response: %s", string(resp))

	gasAfter := ctx.GasMeter().GasConsumed()
	if types.EnableGasVerification {
		require.Equal(t, uint64(0x1a7bb), gasAfter-gasBefore)
	}

	// ensure it is stored properly
	info := keepers.WasmKeeper.GetContractInfo(ctx, gotContractAddr)
	require.NotNil(t, info)
	assert.Equal(t, creator.String(), info.Creator)
	assert.Equal(t, example.CodeID, info.CodeID)
	assert.Equal(t, "demo contract 1", info.Label)

	// verify storage works properly
	prefixStoreKey := types.GetContractStorePrefix(gotContractAddr)
	prefixStore := prefix.NewStore(ctx.KVStore(keepers.WasmKeeper.storeKey), prefixStoreKey)
	configValue := string(prefixStore.Get([]byte("name")))
	assert.Equal(t, initValue, configValue)
	t.Logf("Name value from Smart Contract: %s", configValue)

	exp := []types.ContractCodeHistoryEntry{{
		Operation: types.ContractCodeHistoryOperationTypeInit,
		CodeID:    example.CodeID,
		Updated:   types.NewAbsoluteTxPosition(ctx),
		Msg:       initMsgBz,
	}}
	assert.Equal(t, exp, keepers.WasmKeeper.GetContractHistory(ctx, gotContractAddr))

	// and events emitted
	expEvt := sdk.Events{
		sdk.NewEvent("instantiate",
			sdk.NewAttribute("_contract_address", gotContractAddr.String()), sdk.NewAttribute("code_id", "1")),
		//sdk.NewEvent("wasm",
		//	sdk.NewAttribute("_contract_address", gotContractAddr.String()), sdk.NewAttribute("Let the", "hacking begin")),
	}
	assert.Equal(t, expEvt, em.Events())
}

func TestCosmowrapUpdateName(t *testing.T) {
	ctx, keepers := CreateTestInput(t, false, AvailableCapabilities)
	accKeeper, keeper, bankKeeper := keepers.AccountKeeper, keepers.ContractKeeper, keepers.BankKeeper

	deposit := sdk.NewCoins(sdk.NewInt64Coin("denom", 100000))
	topUp := sdk.NewCoins(sdk.NewInt64Coin("denom", 5000))
	creator := DeterministicAccountAddress(t, 1)
	keepers.Faucet.Fund(ctx, creator, deposit.Add(deposit...)...)
	fred := keepers.Faucet.NewFundedRandomAccount(ctx, topUp...)
	bob := RandomAccountAddress(t)

	contractID, _, err := keeper.Create(ctx, creator, helloWorldWasm, nil)
	require.NoError(t, err)

	initMsgBz := HelloWorldInitMsg{
		name: fred.String(),
	}.GetBytes(t)

	addr, _, err := keepers.ContractKeeper.Instantiate(ctx, contractID, creator, nil, initMsgBz, "demo contract 3", deposit)
	require.NoError(t, err)
	require.Equal(t, "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr", addr.String())

	// ensure bob doesn't exist
	bobAcct := accKeeper.GetAccount(ctx, bob)
	require.Nil(t, bobAcct)

	// ensure funder has reduced balance
	creatorAcct := accKeeper.GetAccount(ctx, creator)
	require.NotNil(t, creatorAcct)
	// we started at 2*deposit, should have spent one above
	assert.Equal(t, deposit, bankKeeper.GetAllBalances(ctx, creatorAcct.GetAddress()))

	// ensure contract has updated balance
	contractAcct := accKeeper.GetAccount(ctx, addr)
	require.NotNil(t, contractAcct)
	assert.Equal(t, deposit, bankKeeper.GetAllBalances(ctx, contractAcct.GetAddress()))

	method := "updateName"
	newName := "Joe"

	updateNameMsgBz := HelloWorldUpdateNameMsg{
		newName: newName,
	}.GetBytes(t)

	// verifier can execute
	em := sdk.NewEventManager()
	_, err = keepers.ContractKeeper.Execute(ctx.WithEventManager(em), addr, fred, updateNameMsgBz, method, topUp)
	require.NoError(t, err)

	// ensure name updated properly in storage
	prefixStoreKey := types.GetContractStorePrefix(addr)
	prefixStore := prefix.NewStore(ctx.KVStore(keepers.WasmKeeper.storeKey), prefixStoreKey)
	configValue := string(prefixStore.Get([]byte("name")))
	assert.Equal(t, newName, configValue)
	t.Logf("Name value from Smart Contract: %s", configValue)
}

func TestCosmowrapSayHello(t *testing.T) {
	ctx, keepers := CreateTestInput(t, false, AvailableCapabilities)
	accKeeper, keeper, bankKeeper := keepers.AccountKeeper, keepers.ContractKeeper, keepers.BankKeeper

	deposit := sdk.NewCoins(sdk.NewInt64Coin("denom", 100000))
	topUp := sdk.NewCoins(sdk.NewInt64Coin("denom", 5000))
	creator := DeterministicAccountAddress(t, 1)
	keepers.Faucet.Fund(ctx, creator, deposit.Add(deposit...)...)
	fred := keepers.Faucet.NewFundedRandomAccount(ctx, topUp...)
	bob := RandomAccountAddress(t)

	contractID, _, err := keeper.Create(ctx, creator, helloWorldWasm, nil)
	require.NoError(t, err)

	initMsgBz := HelloWorldInitMsg{
		name: fred.String(),
	}.GetBytes(t)

	addr, _, err := keepers.ContractKeeper.Instantiate(ctx, contractID, creator, nil, initMsgBz, "demo contract 3", deposit)
	require.NoError(t, err)
	require.Equal(t, "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr", addr.String())

	// ensure bob doesn't exist
	bobAcct := accKeeper.GetAccount(ctx, bob)
	require.Nil(t, bobAcct)

	// ensure funder has reduced balance
	creatorAcct := accKeeper.GetAccount(ctx, creator)
	require.NotNil(t, creatorAcct)
	// we started at 2*deposit, should have spent one above
	assert.Equal(t, deposit, bankKeeper.GetAllBalances(ctx, creatorAcct.GetAddress()))

	// ensure contract has updated balance
	contractAcct := accKeeper.GetAccount(ctx, addr)
	require.NotNil(t, contractAcct)
	assert.Equal(t, deposit, bankKeeper.GetAllBalances(ctx, contractAcct.GetAddress()))

	method := "updateName"
	newName := "Joe"

	updateNameMsgBz := HelloWorldUpdateNameMsg{
		newName: newName,
	}.GetBytes(t)

	// verifier can execute
	em := sdk.NewEventManager()
	_, err = keepers.ContractKeeper.Execute(ctx.WithEventManager(em), addr, fred, updateNameMsgBz, method, topUp)
	require.NoError(t, err)

	// ensure name updated properly in storage
	prefixStoreKey := types.GetContractStorePrefix(addr)
	prefixStore := prefix.NewStore(ctx.KVStore(keepers.WasmKeeper.storeKey), prefixStoreKey)
	configValue := string(prefixStore.Get([]byte("name")))
	assert.Equal(t, newName, configValue)
	//t.Logf("Name value from Smart Contract: %s", configValue)

	method = "sayHello"
	res, err := keepers.ContractKeeper.Execute(ctx.WithEventManager(em), addr, fred, nil, method, nil)
	require.NoError(t, err)
	assert.Equal(t, "Hello from CosmoWrap, Joe", string(res))
	t.Logf("Response: %s", res)
}
