package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"

	evmtypes "github.com/FluxNFTLabs/sdk-go/chain/modules/evm/types"
	chaintypes "github.com/FluxNFTLabs/sdk-go/chain/types"
	chainclient "github.com/FluxNFTLabs/sdk-go/client/chain"
	"github.com/FluxNFTLabs/sdk-go/client/common"
	"github.com/FluxNFTLabs/sdk-go/client/svm"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

	fmt.Println("deploy EVM setter program ...")
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	bz, err := os.ReadFile(dir + "examples/cross_vm/evm_setter/build/compData.json")
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

	//////// DEPLOY SVM PROGRAM ////////
	fmt.Println("deploy SVM setter program ...")
	programBinary, err := os.ReadFile(dir + "/examples/cross_vm/svm_setter/build/svm_setter.so")
	if err != nil {
		panic(err)
	}

	programPrivateKey, err := os.ReadFile(dir + "/examples/cross_vm/svm_setter/build/svm_setter_keypair.json")
	if err != nil {
		panic(err)
	}

	cosmosPrivateKeys := []*ethsecp256k1.PrivKey{
		{Key: ethcommon.Hex2Bytes("88cbead91aee890d27bf06e003ade3d4e952427e88f88d31d61d3ef5e5d54305")},
		{Key: ethcommon.Hex2Bytes("88cbead91aee890d27bf06e003ade3d4e952427e88f88d31d61d3ef5e5d54306")},
		{Key: ethcommon.Hex2Bytes("88cbead91aee890d27bf06e003ade3d4e952427e88f88d31d61d3ef5e5d54307")},
	}
	cosmosAddrs := make([]sdk.AccAddress, len(cosmosPrivateKeys))
	for i, pk := range cosmosPrivateKeys {
		cosmosAddrs[i] = sdk.AccAddress(pk.PubKey().Address().Bytes())
	}

	// create missing cosmos accounts to sign upload message
	msgs := []sdk.Msg{}
	for _, addr := range cosmosAddrs[1:] {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: chainClient.FromAddress().String(),
			ToAddress:   addr.String(),
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("lux", 100000000000000000)),
		})
	}
	_, err = chainClient.SyncBroadcastMsg(msgs...)
	if err != nil {
		panic(err)
	}

	// prepare svm accounts
	ownerSvmPrivKey := ed25519.GenPrivKeyFromSecret([]byte("owner"))
	var programSvmPrivKeyBz []byte
	if err := json.Unmarshal(programPrivateKey, &programSvmPrivKeyBz); err != nil {
		panic(err)
	}

	programSvmPrivKey := &ed25519.PrivKey{Key: programSvmPrivKeyBz}
	programBufferSvmPrivKey := ed25519.GenPrivKeyFromSecret([]byte("programBuffer"))

	// link cosmsos >< svm accounts
	ownerPubkey, _, err := svm.GetOrLinkSvmAccount(chainClient, clientCtx, cosmosPrivateKeys[0], ownerSvmPrivKey, 1000000000000000000)
	if err != nil {
		panic(err)
	}

	programPubkey, _, err := svm.GetOrLinkSvmAccount(chainClient, clientCtx, cosmosPrivateKeys[1], programSvmPrivKey, 0)
	if err != nil {
		panic(err)
	}

	programBufferPubkey, _, err := svm.GetOrLinkSvmAccount(chainClient, clientCtx, cosmosPrivateKeys[2], programBufferSvmPrivKey, 0)
	if err != nil {
		panic(err)
	}

	initAccountMsg := svm.CreateInitAccountsMsg(
		cosmosAddrs,
		len(programBinary),
		ownerPubkey,
		programPubkey,
		programBufferPubkey,
	)

	uploadMsgs, err := svm.CreateProgramUploadMsgs(
		cosmosAddrs,
		ownerPubkey,
		programPubkey,
		programBufferPubkey,
		programBinary,
	)
	if err != nil {
		panic(err)
	}

	signedTx, err := svm.BuildSignedTx(chainClient, []sdk.Msg{initAccountMsg}, cosmosPrivateKeys)
	if err != nil {
		panic(err)
	}

	txBytes, err := chainClient.ClientContext().TxConfig.TxEncoder()(signedTx)
	if err != nil {
		panic(err)
	}

	res, err := chainClient.SyncBroadcastSignedTx(txBytes)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx hash:", res.TxResponse.TxHash)
	fmt.Println("gas used/want:", res.TxResponse.GasUsed, "/", res.TxResponse.GasWanted)

	// upload each parts of the program
	for _, uploadMsg := range uploadMsgs {
		signedTx, err = svm.BuildSignedTx(chainClient, []sdk.Msg{uploadMsg}, cosmosPrivateKeys)
		if err != nil {
			panic(err)
		}

		txBytes, err = chainClient.ClientContext().TxConfig.TxEncoder()(signedTx)
		if err != nil {
			panic(err)
		}

		res, err = chainClient.SyncBroadcastSignedTx(txBytes)
		if err != nil {
			panic(err)
		}

		fmt.Println("tx hash:", res.TxResponse.TxHash)
		fmt.Println("gas used/want:", res.TxResponse.GasUsed, "/", res.TxResponse.GasWanted)
		if res.TxResponse.Code != 0 {
			fmt.Println("err code:", res.TxResponse.Code, ", log:", res.TxResponse.RawLog)
		}
	}
	fmt.Println("program pubkey:", programPubkey.String())
}
