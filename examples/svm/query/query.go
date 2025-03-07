package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "embed"

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
	clientCtx, _, err := chaintypes.NewClientContext(
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

	counterPubkey, _, err := solana.FindProgramAddress([][]byte{
		[]byte("counter"),
	}, programId)
	if err != nil {
		panic(err)
	}

	accountInfo, err := chainClient.GetSvmAccount(context.Background(), counterPubkey.String())
	if err != nil {
		panic(err)
	}

	if len(accountInfo.Account.Data) >= 16 {
		counterValue := binary.LittleEndian.Uint64(accountInfo.Account.Data[8:16])
		fmt.Printf("Counter value: %d\n", counterValue)
	}
}
