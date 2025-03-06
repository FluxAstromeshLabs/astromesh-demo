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
	"github.com/FluxNFTLabs/sdk-go/client/svm"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ethsecp256k1"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
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
		"user2",
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

	// Generate a new account for the counter
	counterPrivKey := ed25519.GenPrivKeyFromSecret([]byte("counter"))
	counterPubkey := solana.PublicKeyFromBytes(counterPrivKey.PubKey().Bytes())

	cosmosPrivateKeys := []*ethsecp256k1.PrivKey{
		{Key: ethcommon.Hex2Bytes("88cbead91aee890d27bf06e003ade3d4e952427e88f88d31d61d3ef5e5d54306")},
	}

	counterPubkey, _, err = svm.GetOrLinkSvmAccount(chainClient, clientCtx, cosmosPrivateKeys[0], counterPrivKey, 0)
	if err != nil {
		panic(err)
	}

	// Check and link fee payer account (using sender's linked SVM account)
	isSvmLinked, feePayerPubkey, err := chainClient.GetSVMAccountLink(context.Background(), senderAddress)
	if err != nil {
		panic(err)
	}
	if !isSvmLinked {
		feePayerPrivKey := ed25519.GenPrivKey() // Good practice: Backup this private key
		res, err := chainClient.LinkSVMAccount(feePayerPrivKey, math.NewIntFromUint64(1_000_000_000_000_000))
		if err != nil {
			panic(err)
		}
		fmt.Println("linked sender to svm address:", base58.Encode(feePayerPrivKey.PubKey().Bytes()), "txHash:", res.TxResponse.TxHash)
		feePayerPubkey = solana.PublicKey(feePayerPrivKey.PubKey().Bytes())
	} else {
		fmt.Println("sender is already linked to svm address:", feePayerPubkey.String())
	}

	// Convert amount to bytes
	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, 5)

	// Anchor discriminators (first 8 bytes of sha256 hash of the instruction name)
	initializeDiscriminator := []byte{175, 175, 109, 31, 13, 152, 155, 237} // "initialize"
	incrementDiscriminator := []byte{127, 205, 142, 147, 115, 108, 45, 143} // "increment"

	txBuilder.AddInstruction(
		solana.NewInstruction(
			programPubkey,
			[]*solana.AccountMeta{
				{PublicKey: counterPubkey, IsSigner: true, IsWritable: true},
				{PublicKey: feePayerPubkey, IsSigner: true, IsWritable: true},
				{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
			},
			initializeDiscriminator,
		),
	)

	// add the increment instruction
	txBuilder.AddInstruction(
		solana.NewInstruction(
			programPubkey,
			[]*solana.AccountMeta{
				{PublicKey: counterPubkey, IsSigner: false, IsWritable: true},
				{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
			},
			append(incrementDiscriminator, amountBytes...),
		),
	)

	tx, err := txBuilder.SetFeePayer(feePayerPubkey).Build()
	if err != nil {
		panic(err)
	}

	svmMsg := svmtypes.ToCosmosMsg([]string{senderAddress.String()}, 1000_000, tx)
	res, err := chainClient.SyncBroadcastMsg(svmMsg)
	if err != nil {
		panic(err)
	}

	fmt.Println("tx hash:", res.TxResponse.TxHash)
	fmt.Println("gas used/want:", res.TxResponse.GasUsed, "/", res.TxResponse.GasWanted)
}
