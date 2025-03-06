package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "embed"

	"cosmossdk.io/math"
	svmtypes "github.com/FluxNFTLabs/sdk-go/chain/modules/svm/types"
	chaintypes "github.com/FluxNFTLabs/sdk-go/chain/types"
	chainclient "github.com/FluxNFTLabs/sdk-go/client/chain"
	"github.com/FluxNFTLabs/sdk-go/client/common"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/gagliardetto/solana-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	privKey []byte
)

func main() {
	network := common.LoadNetwork("local", "")
	kr, err := keyring.New(
		"fluxd",
		"file",
		os.Getenv("HOME")+"/.fluxd",
		strings.NewReader("12345678"),
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

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	privKey, err = os.ReadFile(dir + "/examples/svm/build/counter-contract-keypair.json")
	if err != nil {
		panic(err)
	}

	var privKeyBz []byte
	if err := json.Unmarshal(privKey, &privKeyBz); err != nil {
		panic(err)
	}
	privKey := &ed25519.PrivKey{Key: privKeyBz}
	programId := solana.PublicKeyFromBytes(privKey.PubKey().Bytes())
	programPubkey := programId

	// init chain client
	chainClient, err := chainclient.NewChainClient(
		clientCtx,
		common.OptionGasPrices("500000000lux"),
	)
	if err != nil {
		panic(err)
	}

	txBuilder := solana.NewTransactionBuilder()

	counterPrivKey := ed25519.GenPrivKeyFromSecret([]byte("counter"))
	counterPubkey := solana.PublicKeyFromBytes(counterPrivKey.PubKey().Bytes())

	// Check and link counter account
	isSvmLinked, counterSvmPubkey, err := chainClient.GetSVMAccountLink(context.Background(), senderAddress)
	if err != nil {
		panic(err)
	}
	if !isSvmLinked {
		res, err := chainClient.LinkSVMAccount(counterPrivKey, math.NewInt(0))
		if err != nil {
			panic(err)
		}
		fmt.Printf("linked counter to svm address: %s, txHash: %s\n", counterPubkey.String(), res.TxResponse.TxHash)
		counterPubkey = solana.PublicKeyFromBytes(counterPrivKey.PubKey().Bytes())
	} else {
		fmt.Printf("counter %s is already linked to svm address: %s\n", senderAddress.String(), counterSvmPubkey.String())
		counterPubkey = counterSvmPubkey
	}

	// Convert amount to bytes
	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, 5)

	txBuilder.AddInstruction(
		solana.NewInstruction(
			programPubkey,
			[]*solana.AccountMeta{
				{PublicKey: counterPubkey, IsSigner: false, IsWritable: true},
				{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
			},
			append([]byte{1}, amountBytes...),
		),
	)

	tx, err := txBuilder.Build()
	if err != nil {
		panic(err)
	}

	// Convert ed25519 private keys to Solana format for signing
	counterSolanaKey := solana.PrivateKey(counterPrivKey.Bytes())

	tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(counterPubkey) {
			return &counterSolanaKey
		}
		return nil
	})

	svmMsg := svmtypes.ToCosmosMsg([]string{senderAddress.String()}, 1000_000, tx)
	res, err := chainClient.SyncBroadcastMsg(svmMsg)
	if err != nil {
		panic(err)
	}

	fmt.Println("tx hash:", res.TxResponse.TxHash)
	fmt.Println("gas used/want:", res.TxResponse.GasUsed, "/", res.TxResponse.GasWanted)
}
