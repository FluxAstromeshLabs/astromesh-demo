package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	evmtypes "github.com/FluxNFTLabs/sdk-go/chain/modules/evm/types"
	chaintypes "github.com/FluxNFTLabs/sdk-go/chain/types"
	"github.com/FluxNFTLabs/sdk-go/client/common"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	chainclient "github.com/FluxNFTLabs/sdk-go/client/chain"
)

func main() {
	network := common.LoadNetwork("local", "")
	kr, err := keyring.New(
		"fluxd",
		"file",
		os.Getenv("HOME")+"/.fluxd",
		strings.NewReader("12345678\n"),
		chainclient.GetCryptoCodec(),
	)
	if err != nil {
		panic(err)
	}

	// init grpc connection
	cc, err := grpc.Dial(network.ChainGrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	// init client ctx
	clientCtx, _, err := chaintypes.NewClientContext(
		network.ChainId,
		"user1",
		kr,
	)
	if err != nil {
		panic(err)
	}
	clientCtx = clientCtx.WithGRPCClient(cc)

	evmClient := evmtypes.NewQueryClient(cc)

	// client should know contract's ABI to build this payload
	var compData map[string]interface{}
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	bz, err := os.ReadFile(dir + "/examples/evm/build/compData.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bz, &compData)
	if err != nil {
		panic(err)
	}
	abiBz, err := json.Marshal(compData["abi"].([]interface{}))
	if err != nil {
		panic(err)
	}
	abi, err := abi.JSON(strings.NewReader(string(abiBz)))
	if err != nil {
		panic(err)
	}

	// copy contract address from deploy example
	contractAddress, _ := hex.DecodeString("d2c305699bcda7e79ac26d100b820503f7b15d44")

	// prepare query message, use method get()
	callData, err := abi.Pack("get")
	if err != nil {
		panic(err)
	}

	msg := &evmtypes.ContractQueryRequest{
		Address:  hex.EncodeToString(contractAddress),
		Calldata: callData,
	}

	txResp, err := evmClient.ContractQuery(context.Background(), msg)
	if err != nil {
		panic(err)
	}

	// Print the last byte which contains our counter value
	fmt.Println("Counter value:", new(big.Int).SetBytes(txResp.Output).String())
}
