# astromesh-demo

Install and start the "fluxd" chain, a Cosmos chain with EVM, WASM, and SVM integrated

## Setup chain

There are basically two options:

Option 1: Use Docker

```
./start-docker.sh
```

Option 2: Don't have Docker? Install dependencies and run the binary directly:

```
# install vm libs & fluxd binary
./install.sh

# v0.3.1
fluxd version

# start the chain
fluxd start
```

## Run examples

After bringing up the chain, install golang here as examples are written in go

https://go.dev/doc/install

Now you're ready to run the examples.

This repo contains `counter` contract/program as examples, which supports:
- A `count()` that increases the `counter` by one (counter += 1)
- A `get()` (EVM) that returns the current counter's value

### EVM

Deploy contract

```
yes 12345678 | go run examples/evm/deploy/deploy.go
```

Execute contract to increase the **counter**

```
yes 12345678 | go run examples/evm/execute/execute.go
```

Query contract for counter value

```
yes 12345678 | go run examples/evm/query/query.go
```

Astro Transfer for transfer token from COSMOS to EVM (need to run deploy before transfer)

```
yes 12345678 | go run examples/evm/astro_transfer/example.go
```

### SVM

Deploy program

```
yes 12345678 | go run examples/svm/deploy/deploy.go
```

Execute program to increase the value of **counter** account

```
yes 12345678 | go run examples/svm/execute/execute.go
```

Query program for counter value, as SVM only has accounts, then query the counter account's value

```
yes 12345678 | go run examples/svm/query/query.go
```

Astro Transfer for transfer token from COSMOS to SVM 

```
yes 12345678 | go run examples/svm/astro_transfer/example.go
```

### Cross VM communication: EVM >< SVM

There are two examples contract/program: one for EVM and another from SVM, we can them the `setter` contracts
EVM contract functionalities:

```
setData(s)
    this sets the string value stored in this contract to s
getData() returns string
    this returns the current string stored in the contract
setSvm()
    if svmProgram.getSvmData() == "svm" {
        svmProgram.setData("evm")
    } else {
        svmProgram.setData("svm")
    }
```

Similarly, SVM has same functionalities:

```
setData(s)
    this sets the string value stored in this contract to s
getData() returns string
    this returns the current string stored in the contract
setEvm()
    if evmContract.getEvmData() == "evm" {
        evmContract.setData("svm")
    } else {
        evmContract.setData("evm")
    }
```

Calling `setSvm()`, EVM contract queries the string from SVM program and determine which value should be set for SVM

To run examples:

1. Deploy the 2 contracts:

Before deploy, you can tweak something to check deeper if need

```
yes 12345678 | go run examples/cross_vm/deploy/deploy.go
```

2. Invoke contracts functions:

```
# invoke setSvm() from EVM contract
yes 12345678 | go run examples/cross_vm/execute/evm_to_svm.go

# invoke setEvm() from SVM contract
yes 12345678 | go run examples/cross_vm/execute/svm_to_evm.go
```

Each example also does a query after execution to double check if the value correctly set

## How to compile contract/programs

This guide provides steps to build manually without installing dependencies. We will provide guides using CLI tools later.

### EVM

1. Visit [Remix IDE](https://remix.ethereum.org/), and create a new contract, click "File Explorer" > "Contracts" > "Create New File"

2. Copy contract at `examples/evm/contract/counter.sol`. You can modify it at this step if needed

3. Navigate to the Solidity Compiler tab on the left panel > Click "Compile" > Click "Compilation Details" > Click "Download" (in the top right panel)

4. After download, move it to `examples/evm/build/compData.json`

Note: In the example, we've already built the contract so you don't need to follow any of above steps unless you want to try another

### SVM

1. Visit [Solana Playground](https://beta.solpg.io/), and create a new project:

- Click the "+" button.

- Name the project counter

- Select the Anchor framework.

Copy sample program at `examples/svm/counter_program/lib.rs` into `src/lib.rs` file. You can modify it at this step if needed

2. Navigate to the Build & Deploy tab on the left panel.

For initial build:

- Click "Build". This will generate a new keypair and update the program ID.

For subsequent builds:

- To generate a new program ID, click on "Program ID", then "New", and select "Generate" before building.

3. Download the artifacts:

Keypair: Navigate to "Program ID" and click "Export". Copy the keypair into `examples/svm/build/counter-contract-keypair.json`

Binary: Navigate to "Program Binary" and click "Export". Copy the binary into `examples/svm/build/counter-contract.so`

IDL: Navigate to "IDL" and click "Export". We don't use IDL in examples for now but will instruct how to use it with go later

Note: In the example, we've already built the program so you don't need to follow any of above steps unless you want to try another
