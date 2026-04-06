package simplestorage

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:embed abi/SimpleStorage.json
var simpleStorageABI []byte

type BesuAdapter struct {
	client   *ethclient.Client
	contract *bind.BoundContract
	auth     *bind.TransactOpts
}

func NewBesuAdapter(rpcURL, contractAddr, privKeyHex string) (*BesuAdapter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial besu: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("get chain id: %w", err)
	}

	parsedABI, err := abi.JSON(bytes.NewReader(simpleStorageABI))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("parse abi: %w", err)
	}

	address := common.HexToAddress(contractAddr)
	contract := bind.NewBoundContract(address, parsedABI, client, client, client)

	privKey, err := crypto.ToECDSA(common.FromHex(privKeyHex))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("load private key: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainID)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("create transactor: %w", err)
	}

	return &BesuAdapter{client: client, contract: contract, auth: auth}, nil
}

func (b *BesuAdapter) Close() {
	b.client.Close()
}

func (b *BesuAdapter) SetValue(ctx context.Context, value *big.Int) (string, error) {
	writeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	b.auth.Context = writeCtx

	tx, err := b.contract.Transact(b.auth, "set", value)
	if err != nil {
		return "", fmt.Errorf("transact set: %w", err)
	}

	_, err = bind.WaitMined(writeCtx, b.client, tx)
	if err != nil {
		return "", fmt.Errorf("wait mined: %w", err)
	}

	return tx.Hash().Hex(), nil
}

func (b *BesuAdapter) GetValue(ctx context.Context) (*big.Int, error) {
	readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := &bind.CallOpts{Context: readCtx}

	var output []interface{}
	if err := b.contract.Call(opts, &output, "get"); err != nil {
		return nil, fmt.Errorf("call get: %w", err)
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("empty output from get")
	}

	v, ok := output[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected output type %T", output[0])
	}

	return v, nil
}
