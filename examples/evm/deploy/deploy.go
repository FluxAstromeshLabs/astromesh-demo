package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/accounts/abi"

	astromeshtypes "github.com/FluxNFTLabs/sdk-go/chain/modules/astromesh/types"
	evmtypes "github.com/FluxNFTLabs/sdk-go/chain/modules/evm/types"
	chaintypes "github.com/FluxNFTLabs/sdk-go/chain/types"
	chainclient "github.com/FluxNFTLabs/sdk-go/client/chain"
	"github.com/FluxNFTLabs/sdk-go/client/common"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	clientCtx, senderAddress, err := chaintypes.NewClientContext(
		network.ChainId,
		"user2",
		kr,
	)
	if err != nil {
		panic(err)
	}

	clientCtx = clientCtx.WithGRPCClient(cc)
	// init chain client
	chainClient, err := chainclient.NewChainClient(
		clientCtx,
		common.OptionGasPrices("500000000lux"),
	)
	if err != nil {
		panic(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	bz, err := os.ReadFile(dir + "/examples/evm/build/compData.json")
	if err != nil {
		panic(err)
	}
	var compData map[string]interface{}
	err = json.Unmarshal(bz, &compData)
	if err != nil {
		panic(err)
	}
	bytecode, err := hex.DecodeString(compData["bytecode"].(map[string]interface{})["object"].(string))
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
	callData, err := abi.Pack("")
	if err != nil {
		panic(err)
	}

	// prepare deploy msg
	msg := &evmtypes.MsgDeployContract{
		Sender:   senderAddress.String(),
		Bytecode: bytecode,
		Calldata: callData,
	}

	txResp, err := chainClient.SyncBroadcastMsg(msg)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx hash:", txResp.TxResponse.TxHash)
	fmt.Println("gas used/want:", txResp.TxResponse.GasUsed, "/", txResp.TxResponse.GasWanted)
	hexResp, err := hex.DecodeString(txResp.TxResponse.Data)
	if err != nil {
		panic(err)
	}

	// decode result to get contract address
	var txData1 sdk.TxMsgData
	if err := txData1.Unmarshal(hexResp); err != nil {
		panic(err)
	}
	var dcr evmtypes.MsgDeployContractResponse
	if err := dcr.Unmarshal(txData1.MsgResponses[0].Value); err != nil {
		panic(err)
	}
	fmt.Println("contract owner:", senderAddress.String())
	fmt.Println("contract address:", hex.EncodeToString(dcr.ContractAddress))

	// prepare transfer msg
	msg2 := &astromeshtypes.MsgAstroTransfer{
		Sender:   senderAddress.String(),
		Receiver: senderAddress.String(),
		SrcPlane: astromeshtypes.Plane_COSMOS,
		DstPlane: astromeshtypes.Plane_EVM,
		Coin: sdk.Coin{
			Denom:  "lux",
			Amount: math.NewIntFromUint64(1000000000000000000), // 1^18
		},
	}
	txResp, err = chainClient.SyncBroadcastMsg(msg2)
	if err != nil {
		panic(err)
	}
	fmt.Println("astro transfer tx hash:", txResp.TxResponse.TxHash)
	fmt.Println("gas used/want:", txResp.TxResponse.GasUsed, "/", txResp.TxResponse.GasWanted)
}
