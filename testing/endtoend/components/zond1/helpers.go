package zond1

import (
	"context"
	"math/big"
	"time"

	"github.com/theQRL/go-zond/accounts/keystore"
	"github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/zondclient"
	"github.com/theQRL/qrysm/v4/testing/endtoend/params"
	e2etypes "github.com/theQRL/qrysm/v4/testing/endtoend/types"
)

// NetworkId is the ID of the ZOND1 chain.
const NetworkId = 1337

// KeystorePassword is the password used to decrypt ZOND1 keystores.
const KeystorePassword = "password"

const minerPasswordFile = "password.txt"

// TODO(rgeraldes24) - might have to change the key
const minerFile = "UTC--2021-12-22T19-14-08.590377700Z--878705ba3f8bc32fcf7f4caa1a35e72af65cf766"
const timeGapPerMiningTX = 250 * time.Millisecond

var _ e2etypes.ComponentRunner = (*NodeSet)(nil)
var _ e2etypes.MultipleComponentRunners = (*NodeSet)(nil)
var _ e2etypes.MultipleComponentRunners = (*ProxySet)(nil)
var _ e2etypes.ComponentRunner = (*Miner)(nil)
var _ e2etypes.ComponentRunner = (*Node)(nil)
var _ e2etypes.EngineProxy = (*Proxy)(nil)

// WaitForBlocks waits for a certain amount of blocks to be mined by the ZOND1 chain before returning.
func WaitForBlocks(web3 *zondclient.Client, key *keystore.Key, blocksToWait uint64) error {
	nonce, err := web3.PendingNonceAt(context.Background(), key.Address)
	if err != nil {
		return err
	}
	chainID, err := web3.NetworkID(context.Background())
	if err != nil {
		return err
	}
	block, err := web3.BlockByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	finishBlock := block.NumberU64() + blocksToWait

	for block.NumberU64() <= finishBlock {
		gasPrice, err := web3.SuggestGasPrice(context.Background())
		if err != nil {
			return err
		}
		spamTX := types.NewTransaction(nonce, key.Address, big.NewInt(0), params.SpamTxGasLimit, gasPrice, []byte{})
		signed, err := types.SignTx(spamTX, types.NewEIP155Signer(chainID), key.Dilithium)
		if err != nil {
			return err
		}
		if err = web3.SendTransaction(context.Background(), signed); err != nil {
			return err
		}
		nonce++
		time.Sleep(timeGapPerMiningTX)
		block, err = web3.BlockByNumber(context.Background(), nil)
		if err != nil {
			return err
		}
	}
	return nil
}