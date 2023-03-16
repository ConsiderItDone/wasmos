# CosmoWrap


**CosmoWrap is a Smart contract system based on WebAssembly running on Cosmos Network with the following key features**:

- Allows development of smart contracts using Rust, Go or AssembyScript
- Improved developer toolchain with focus on enhanced integration testing
    - Out-of-box Integration Test are provided (CosmWasm only provides unit tests)
- Automated binding process between Cosmos modules and smart contacts
    - Proto File ([Protocol Buffer](https://github.com/protocolbuffers/protobuf)) based scaffolding for smart contracts
- Ability to call any entry-point function in Smart Contracts with improved code readability and organization
    - CosmWasm has a single [execute()](https://github.com/CosmWasm/cosmwasm/blob/main/contracts/staking/src/contract.rs#L60) entry point with function execution derived from message in the  argument.
    - CosmoWrap allow to call any function directly
- Improved sub-messaging invocation with ability to call Cosmos Chain functions directly
    - improved in function response handing
    - improved execution flow control
    - [CosmWasm example response message](https://github.com/CosmWasm/cosmwasm/blob/main/contracts/staking/src/contract.rs#L176)
- Provides access to IBC, Cosmos Staking and Cosmos key-value storage (get and set data on-chain)


![](https://i.imgur.com/SA4t7TZ.png)

## Development

See [Hello World](https://github.com/ConsiderItDone/cosmowrap-hello-world-as/) smart contract example.


Take a look at test file [`x/wasm/keeper/cosmowrap_test.go`](x/wasm/keeper/cosmowrap_test.go.go)


```shell
go test -v -run '^TestCosmowrap' ./x/wasm/keeper
```