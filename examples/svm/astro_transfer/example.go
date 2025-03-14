package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gagliardetto/solana-go"

	"cosmossdk.io/math"
	astromeshtypes "github.com/FluxNFTLabs/sdk-go/chain/modules/astromesh/types"
	chaintypes "github.com/FluxNFTLabs/sdk-go/chain/types"
	chainclient "github.com/FluxNFTLabs/sdk-go/client/chain"
	"github.com/FluxNFTLabs/sdk-go/client/common"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// init chain client
	chainClient, err := chainclient.NewChainClient(
		clientCtx,
		common.OptionGasPrices("500000000lux"),
	)
	if err != nil {
		fmt.Println(err)
	}

	// prepare tx msg
	if err != nil {
		panic(err)
	}

	astromeshClient := astromeshtypes.NewQueryClient(cc)
	clientCtx = clientCtx.WithGRPCClient(cc)

	denomLink, err := astromeshClient.DenomLink(context.Background(), &astromeshtypes.QueryDenomLinkRequest{
		SrcPlane: astromeshtypes.Plane_COSMOS,
		DstPlane: astromeshtypes.Plane_SVM,
		SrcAddr:  "lux",
	})
	if err != nil {
		panic(err)
	}

	luxDenomBz, _ := hex.DecodeString(denomLink.DstAddr)
	luxDenom := solana.PublicKeyFromBytes(luxDenomBz)

	balanceResp, err := astromeshClient.Balance(context.Background(), &astromeshtypes.BalanceRequest{
		Plane:   astromeshtypes.Plane_SVM.String(),
		Denom:   luxDenom.String(),
		Address: senderAddress.String(),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("balance before transfer:", balanceResp.Amount)

	// prepare tx msg
	msg1 := &astromeshtypes.MsgAstroTransfer{
		Sender:   senderAddress.String(),
		Receiver: senderAddress.String(),
		SrcPlane: astromeshtypes.Plane_COSMOS,
		DstPlane: astromeshtypes.Plane_SVM,
		Coin: sdk.Coin{
			Denom:  "lux",
			Amount: math.NewIntFromUint64(1000000000000000000), // 1^18
		},
	}
	txResp, err := chainClient.SyncBroadcastMsg(msg1)
	if err != nil {
		panic(err)
	}
	fmt.Println("resp:", txResp.TxResponse.TxHash)
	fmt.Println("gas used/want:", txResp.TxResponse.GasUsed, "/", txResp.TxResponse.GasWanted)

	balanceResp, err = astromeshClient.Balance(context.Background(), &astromeshtypes.BalanceRequest{
		Plane:   astromeshtypes.Plane_SVM.String(),
		Denom:   luxDenom.String(),
		Address: senderAddress.String(),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("balance after transfer:", balanceResp.Amount)
}
