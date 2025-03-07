package main

import (
	"context"
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

	// init chain client
	chainClient, err := chainclient.NewChainClient(
		clientCtx,
		common.OptionGasPrices("500000000lux"),
	)
	if err != nil {
		panic(err)
	}

	txBuilder := solana.NewTransactionBuilder()

	counterPubkey, _, err := solana.FindProgramAddress([][]byte{
		[]byte("counter"),
	}, programId)
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

	// Anchor discriminators (first 8 bytes of sha256 hash of "global:count")
	countDiscriminator := []byte{214, 3, 93, 57, 210, 192, 181, 206} // "global:count"

	// add the count instruction
	txBuilder.AddInstruction(
		solana.NewInstruction(
			programId,
			[]*solana.AccountMeta{
				{PublicKey: counterPubkey, IsSigner: false, IsWritable: true},
				{PublicKey: feePayerPubkey, IsSigner: true, IsWritable: true},
				{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
			},
			countDiscriminator,
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
