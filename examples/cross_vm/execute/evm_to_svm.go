package main

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/btcutil/base58"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/gagliardetto/solana-go"

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
	clientCtx, senderAddress, err := chaintypes.NewClientContext(
		network.ChainId,
		"user1",
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

	// client should know contract's ABI to build this payload
	var compData map[string]interface{}
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	bz, err := os.ReadFile(dir + "/examples/cross_vm/evm_setter/build/compData.json")
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

	isSvmLinked, svmPubkey, err := chainClient.GetSVMAccountLink(context.Background(), senderAddress)
	if err != nil {
		panic(err)
	}

	if !isSvmLinked {
		svmKey := ed25519.GenPrivKey() // Good practice: Backup this private key
		res, err := chainClient.LinkSVMAccount(svmKey, math.NewIntFromUint64(1000_000_000_000))
		if err != nil {
			panic(err)
		}
		fmt.Println("linked sender to svm address:", base58.Encode(svmKey.PubKey().Bytes()), "txHash:", res.TxResponse.TxHash)
	} else {
		fmt.Println("sender is already linked to svm address:", svmPubkey.String())
	}

	// TODO: copy contract address from deploy example
	contractAddress, _ := hex.DecodeString("ab6b4d064c968eca87f775d2493a222987052bc0")
	callData, err := abi.Pack("setSvm", svmPubkey.String())
	if err != nil {
		panic(err)
	}

	// prepare message to invoke count()
	msg := &evmtypes.MsgExecuteContract{
		Sender:          senderAddress.String(),
		ContractAddress: contractAddress,
		Calldata:        callData,
	}

	txResp, err := chainClient.SyncBroadcastMsg(msg)
	if err != nil {
		panic(err)
	}
	fmt.Println("txHash:", txResp.TxResponse.TxHash)
	fmt.Println("gas used/want:", txResp.TxResponse.GasUsed, "/", txResp.TxResponse.GasWanted)

	// Double check SVM's string value
	programId := solana.MustPublicKeyFromBase58("83zfZYacFrGq5eBnnp6EQPxapcpjpxdjAKpLavqtSJ32")
	dataPDA, _, err := solana.FindProgramAddress([][]byte{[]byte("data")}, programId)
	if err != nil {
		panic(err)
	}

	acc, err := chainClient.GetSvmAccount(context.Background(), dataPDA.String())
	if err != nil {
		panic(err)
	}

	anchorData := acc.Account.Data[8:]
	len := binary.LittleEndian.Uint32(anchorData[:4])
	strInfo := anchorData[4 : 4+len]
	fmt.Println("SVM string value:", string(strInfo))
}
