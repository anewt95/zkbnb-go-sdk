package client

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bnb-chain/zkbnb-go-sdk/accounts"
	"github.com/bnb-chain/zkbnb-go-sdk/signer"
	"github.com/bnb-chain/zkbnb-go-sdk/txutils"
	"github.com/bnb-chain/zkbnb-go-sdk/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

const defaultExpireTime = time.Minute * 10

var (
	dialer = &net.Dialer{
		Timeout:   1 * time.Second,
		KeepAlive: 60 * time.Second,
	}
	transport = &http.Transport{
		DialContext:         dialer.DialContext,
		MaxConnsPerHost:     1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	HttpClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
)

type l2Client struct {
	endpoint   string
	privateKey string
	chainId    uint64
	l1Signer   signer.L1Signer
	keyManager accounts.KeyManager
}

func (c *l2Client) KeyManager() accounts.KeyManager {
	return c.keyManager
}

func (c *l2Client) GetCurrentHeight() (int64, error) {
	resp, err := HttpClient.Get(c.endpoint + fmt.Sprintf("/api/v1/currentHeight"))
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf(string(body))
	}
	result := &types.CurrentHeight{}
	if err := json.Unmarshal(body, result); err != nil {
		return -1, err
	}
	return result.Height, nil
}

func (c *l2Client) GetTxsByAccountPk(accountPk string, offset, limit uint32, options ...GetTxOptionFunc) (total uint32, txs []*types.Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	path := fmt.Sprintf("/api/v1/accountTxs?by=account_pk&value=%s&offset=%d&limit=%d", accountPk, offset, limit)
	if len(opt.Types) > 0 {
		txTypes, _ := json.Marshal(opt.Types)
		path += fmt.Sprintf("&types=%s", string(txTypes))
	}

	resp, err := HttpClient.Get(c.endpoint + path)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	result := &types.Txs{}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, nil, err
	}
	return result.Total, result.Txs, nil
}

func (c *l2Client) GetTxsByL1Address(l1Address string, offset, limit uint32, options ...GetTxOptionFunc) (total uint32, txs []*types.Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	path := fmt.Sprintf("/api/v1/accountTxs?by=l1_address&value=%s&offset=%d&limit=%d", l1Address, offset, limit)
	if len(opt.Types) > 0 {
		txTypes, _ := json.Marshal(opt.Types)
		path += fmt.Sprintf("&types=%s", string(txTypes))
	}

	resp, err := HttpClient.Get(c.endpoint + path)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	result := &types.Txs{}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, nil, err
	}
	return result.Total, result.Txs, nil
}

func (c *l2Client) GetTxs(offset, limit uint32) (total uint32, txs []*types.Tx, err error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/txs?offset=%d&limit=%d", offset, limit))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	result := &types.Txs{}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, nil, err
	}
	return result.Total, result.Txs, nil
}

func (c *l2Client) GetTxsByAccountIndex(accountIndex int64, offset, limit uint32, options ...GetTxOptionFunc) (total uint32, txs []*types.Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	path := fmt.Sprintf("/api/v1/accountTxs?by=account_index&value=%d&offset=%d&limit=%d", accountIndex, offset, limit)
	if len(opt.Types) > 0 {
		txTypes, _ := json.Marshal(opt.Types)
		path += fmt.Sprintf("&types=%s", string(txTypes))
	}

	resp, err := HttpClient.Get(c.endpoint + path)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	result := &types.Txs{}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, nil, err
	}
	return result.Total, result.Txs, nil
}

func (c *l2Client) Search(keyword string) (*types.Search, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/search?keyword=%s", keyword))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Search{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetAccounts(offset, limit uint32) (*types.Accounts, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/accounts?offset=%d&limit=%d", offset, limit))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Accounts{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetGasFeeAssets() (*types.GasFeeAssets, error) {
	resp, err := HttpClient.Get(c.endpoint + "/api/v1/gasFeeAssets")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.GasFeeAssets{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetGasFee(assetId int64, txType int) (*big.Int, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/gasFee?asset_id=%d&tx_type=%d", assetId, txType))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.GasFee{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	var price big.Int
	price.SetString(result.GasFee, 10)
	return &price, nil
}

func (c *l2Client) GetAssetById(id uint32) (*types.Asset, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/asset?by=id&value=%d", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Asset{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetAssetBySymbol(symbol string) (*types.Asset, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/asset?by=symbol&value=%s", symbol))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Asset{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetAssets(offset, limit uint32) (*types.Assets, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/assets?offset=%d&limit=%d", offset, limit))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Assets{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetLayer2BasicInfo() (*types.Layer2BasicInfo, error) {
	resp, err := HttpClient.Get(c.endpoint + "/api/v1/layer2BasicInfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Layer2BasicInfo{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetRollbacks(fromBlockHeight, offset, limit int64) (total uint32, rollbacks []*types.Rollback, err error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/rollbacks?from_block_height=%d&limit=%d&offset=%d", fromBlockHeight, limit, offset))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	result := &types.Rollbacks{}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, nil, err
	}
	return result.Total, result.Rollbacks, err
}

func (c *l2Client) GetBlockByCommitment(blockCommitment string) (*types.Block, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/block?by=commitment&value=%s", blockCommitment))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Block{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetAccountByIndex(accountIndex int64) (*types.Account, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/account?by=index&value=%d", accountIndex))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Account{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetAccountByPk(accountPk string) (*types.Account, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/account?by=pk&value=%s", accountPk))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Account{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetTx(hash string) (*types.EnrichedTx, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/tx?hash=%s", hash))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	txResp := &types.EnrichedTx{}
	if err := json.Unmarshal(body, txResp); err != nil {
		return nil, err
	}
	return txResp, nil
}

func (c *l2Client) GetPendingTxs(offset, limit uint32) (total uint32, txs []*types.Tx, err error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/pendingTxs?offset=%d&limit=%d", offset, limit))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	txsResp := &types.Txs{}
	if err := json.Unmarshal(body, txsResp); err != nil {
		return 0, nil, err
	}
	return txsResp.Total, txsResp.Txs, nil
}

func (c *l2Client) GetPendingTxsByL1Address(l1Address string, options ...GetTxOptionFunc) (total uint32, txs []*types.Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	path := "/api/v1/accountPendingTxs?by=l1_address&value=" + l1Address
	if len(opt.Types) > 0 {
		txTypes, _ := json.Marshal(opt.Types)
		path += fmt.Sprintf("&types=%s", string(txTypes))
	}

	resp, err := HttpClient.Get(c.endpoint + path)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	txsResp := &types.Txs{}
	if err := json.Unmarshal(body, txsResp); err != nil {
		return 0, nil, err
	}
	return txsResp.Total, txsResp.Txs, nil
}

func (c *l2Client) GetExecutedTxs(offset, limit uint32, options ...GetTxOptionFunc) (total uint32, txs []*types.Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	path := fmt.Sprintf("/api/v1/executedTxs?offset=%d&limit=%d", offset, limit)
	if len(opt.FromHash) > 0 {
		path += fmt.Sprintf("&from_hash=%s", opt.FromHash)
	}

	resp, err := HttpClient.Get(c.endpoint + path)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	txsResp := &types.Txs{}
	if err := json.Unmarshal(body, txsResp); err != nil {
		return 0, nil, err
	}
	return txsResp.Total, txsResp.Txs, nil
}

func (c *l2Client) GetAccountByL1Address(l1Address string) (*types.Account, error) {
	resp, err := HttpClient.Get(c.endpoint + "/api/v1/account?by=l1_address&value=" + l1Address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	account := &types.Account{}
	if err := json.Unmarshal(body, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (c *l2Client) GetNextNonce(accountIdx int64) (int64, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/nextNonce?account_index=%d", accountIdx))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(string(body))
	}
	result := &types.NextNonce{}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, err
	}
	return int64(result.Nonce), nil
}

func (c *l2Client) GetTxsByBlockHeight(blockHeight uint32) ([]*types.Tx, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/blockTxs?by=block_height&value=%d", blockHeight))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.Txs{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result.Txs, nil
}

func (c *l2Client) GetMaxOfferId(accountIndex int64) (uint64, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/maxOfferId?account_index=%d", accountIndex))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(string(body))
	}
	result := &types.MaxOfferId{}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, err
	}
	return result.OfferId, nil
}

func (c *l2Client) GetBlockByHeight(blockHeight int64) (*types.Block, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/block?by=height&value=%d", blockHeight))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	res := &types.Block{}
	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *l2Client) GetBlocks(offset, limit int64) (uint32, []*types.Block, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/blocks?limit=%d&offset=%d", limit, offset))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf(string(body))
	}
	res := &types.Blocks{}
	if err := json.Unmarshal(body, res); err != nil {
		return 0, nil, err
	}
	return res.Total, res.Blocks, nil
}

func (c *l2Client) GetGasAccount() (*types.GasAccount, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/gasAccount"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	res := &types.GasAccount{}
	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *l2Client) GetNftsByAccountIndex(accountIndex, offset, limit int64) (*types.Nfts, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/accountNfts?by=account_index&value=%d&limit=%d&offset=%d", accountIndex, limit, offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	res := &types.Nfts{}
	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *l2Client) getL2SignatureBody(txType uint32, txInfo string) (string, error) {
	resp, err := HttpClient.PostForm(c.endpoint+"/api/v1/l2Signature",
		url.Values{"tx_type": {strconv.Itoa(int(txType))}, "tx_info": {txInfo}, "tx_signature": {"-"}})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(string(body))
	}
	res := &types.SignBody{}
	if err := json.Unmarshal(body, res); err != nil {
		return "", err
	}
	return res.SignBody, nil
}

func (c *l2Client) GetMaxCollectionId(accountIndex int64) (*types.MaxCollectionId, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/maxCollectionId?account_index=%d", accountIndex))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.MaxCollectionId{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) GetNftByTxHash(txHash string) (*types.NftIndex, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/getNftByTxHash?tx_hash=%s", txHash))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	result := &types.NftIndex{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *l2Client) UpdateNftByIndex(nft *types.UpdateNftReq, signatureList ...string) (*types.Mutable, error) {
	if c.keyManager == nil {
		return nil, fmt.Errorf("key manager is nil")
	}
	if nft.AccountIndex == 0 {
		l2Account, err := c.GetAccountByPk(hex.EncodeToString(c.keyManager.PubKey().Bytes()))
		if err != nil {
			return nil, err
		}
		nft.AccountIndex = l2Account.Index
	}
	if nft.Nonce == 0 {
		nonce, err := c.GetNftNextNonce(nft.NftIndex)
		if err != nil {
			return nil, err
		}
		nft.Nonce = nonce
	}
	txInfo, err := c.constructUpdateNFTTransaction(nft, nil)
	if err != nil {
		return nil, err
	}
	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return nil, err
	}
	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return nil, err
	}
	resp, err := HttpClient.PostForm(c.endpoint+"/api/v1/updateNftByIndex",
		url.Values{"tx_info": {string(txInfoBytes)}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}
	res := &types.Mutable{}
	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *l2Client) GetNftNextNonce(nftIndex int64) (int64, error) {
	resp, err := HttpClient.Get(c.endpoint +
		fmt.Sprintf("/api/v1/nftNextNonce?nft_index=%d", nftIndex))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(string(body))
	}
	result := &types.NextNonce{}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, err
	}
	return int64(result.Nonce), nil
}

func (c *l2Client) SendRawTx(txType uint32, txInfo string) (string, error) {
	resp, err := HttpClient.PostForm(c.endpoint+"/api/v1/sendTx",
		url.Values{"tx_type": {strconv.Itoa(int(txType))}, "tx_info": {txInfo}})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(string(body))
	}
	res := &types.TxHash{}
	if err := json.Unmarshal(body, res); err != nil {
		return "", err
	}
	return res.TxHash, nil
}

func (c *l2Client) ChangePubKey(tx *types.ChangePubKeyReq, ops *types.TransactOpts, signatureList ...string) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}
	account, err := c.GetAccountByL1Address(tx.L1Address)
	if err != nil {
		return "", err
	}
	ops.FromAccountIndex = account.Index
	txInfo, err := c.constructChangePubKeyTransaction(tx, ops)
	if err != nil {
		return "", err
	}
	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return "", err
	}
	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) MintNft(tx *types.MintNftTxReq, ops *types.TransactOpts, signatureList ...string) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}

	txInfo, err := c.constructMintNftTransaction(tx, ops)
	if err != nil {
		return "", err
	}

	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return "", err
	}
	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) CreateCollection(tx *types.CreateCollectionTxReq, ops *types.TransactOpts, signatureList ...string) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}

	txInfo, err := c.constructCreateCollectionTransaction(tx, ops)
	if err != nil {
		return "", err
	}
	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return "", err
	}
	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) CancelOffer(tx *types.CancelOfferTxReq, ops *types.TransactOpts, signatureList ...string) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}

	txInfo, err := c.constructCancelOfferTransaction(tx, ops)
	if err != nil {
		return "", err
	}

	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return "", err
	}
	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) AtomicMatch(tx *types.AtomicMatchTxReq, ops *types.TransactOpts) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}
	if ops == nil {
		ops = new(types.TransactOpts)
	}
	txInfo, err := c.constructAtomicMatchTransaction(tx, ops)
	if err != nil {
		return "", err
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) WithdrawNft(tx *types.WithdrawNftTxReq, ops *types.TransactOpts, signatureList ...string) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}

	txInfo, err := c.constructWithdrawNftTransaction(tx, ops)
	if err != nil {
		return "", err
	}
	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return "", err
	}

	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) TransferNft(tx *types.TransferNftTxReq, ops *types.TransactOpts, signatureList ...string) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}

	txInfo, err := c.constructTransferNftTransaction(tx, ops)
	if err != nil {
		return "", err
	}
	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return "", err
	}
	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) Withdraw(tx *types.WithdrawTxReq, ops *types.TransactOpts, signatureList ...string) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}

	txInfo, err := c.constructWithdrawTransaction(tx, ops)
	if err != nil {
		return "", err
	}
	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return "", err
	}

	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) Transfer(tx *types.TransferTxReq, ops *types.TransactOpts, signatureList ...string) (string, error) {
	if c.keyManager == nil {
		return "", fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}

	txInfo, err := c.constructTransferTransaction(tx, ops)
	if err != nil {
		return "", err
	}
	signature, err := c.generateSignature(txInfo, signatureList)
	if err != nil {
		return "", err
	}
	txInfo.L1Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}
	return c.SendRawTx(uint32(txInfo.GetTxType()), string(txInfoBytes))
}

func (c *l2Client) fullFillToAddrOps(ops *types.TransactOpts, to string) (*types.TransactOpts, error) {
	toAccount, err := c.GetAccountByL1Address(to)
	if err != nil {
		return nil, err
	}
	ops.ToAccountIndex = toAccount.Index
	ops.ToAccountAddress = toAccount.L1Address
	return ops, nil
}

func (c *l2Client) fullFillDefaultOps(ops *types.TransactOpts) (*types.TransactOpts, error) {
	if ops == nil {
		ops = new(types.TransactOpts)
	}
	if ops.GasAccountIndex == 0 {
		gasAccount, err := c.GetGasAccount()
		if err != nil {
			return nil, err
		}
		if gasAccount.Index == 0 {
			return nil, fmt.Errorf("get gas account error, gas account index is %d", gasAccount.Index)
		}
		ops.GasAccountIndex = gasAccount.Index
	}
	if ops.ExpiredAt == 0 {
		ops.ExpiredAt = time.Now().Add(defaultExpireTime).UnixMilli()
	}
	if ops.FromAccountIndex == 0 {
		l2Account, err := c.GetAccountByPk(hex.EncodeToString(c.keyManager.PubKey().Bytes()))
		if err != nil {
			return nil, err
		}
		ops.FromAccountIndex = l2Account.Index
	}
	if ops.Nonce == 0 {
		nonce, err := c.GetNextNonce(ops.FromAccountIndex)
		if err != nil {
			return nil, err
		}
		ops.Nonce = nonce
	}
	if len(ops.CallDataHash) == 0 {
		hFunc := mimc.NewMiMC()
		ops.CallDataHash = hFunc.Sum([]byte(ops.CallData))
	}
	if ops.GasFeeAssetAmount == nil {
		gas, err := c.GetGasFee(ops.GasFeeAssetId, ops.TxType)
		if err != nil {
			return nil, err
		}
		ops.GasFeeAssetAmount = gas
	}
	return ops, nil
}

func (c *l2Client) fullFillChangePubKeyOps(ops *types.TransactOpts) (*types.TransactOpts, error) {
	ops.TxType = txtypes.TxTypeChangePubKey
	if ops.GasAccountIndex == 0 {
		gasAccount, err := c.GetGasAccount()
		if err != nil {
			return nil, err
		}
		if gasAccount.Index == 0 {
			return nil, fmt.Errorf("get gas account error, gas account index is %d", gasAccount.Index)
		}
		ops.GasAccountIndex = gasAccount.Index
	}
	if ops.ExpiredAt == 0 {
		ops.ExpiredAt = time.Now().Add(defaultExpireTime).UnixMilli()
	}
	if ops.Nonce == 0 {
		nonce, err := c.GetNextNonce(ops.FromAccountIndex)
		if err != nil {
			return nil, err
		}
		ops.Nonce = nonce
	}
	if len(ops.CallDataHash) == 0 {
		hFunc := mimc.NewMiMC()
		ops.CallDataHash = hFunc.Sum([]byte(ops.CallData))
	}
	if ops.GasFeeAssetAmount == nil {
		gas, err := c.GetGasFee(ops.GasFeeAssetId, ops.TxType)
		if err != nil {
			return nil, err
		}
		ops.GasFeeAssetAmount = gas
	}
	return ops, nil
}

func (c *l2Client) generateSignature(txInfo txtypes.TxInfo, signatureList []string) (string, error) {
	if len(signatureList) == 0 {
		if c.l1Signer == nil {
			return "", errors.New("privateKey has not been initialized correctly, signature is expected to be passed instead")
		}

		signBody := txInfo.GetL1SignatureBody()
		signHex, err := c.l1Signer.Sign(signBody)
		if err != nil {
			return "", err
		}
		return signHex, nil
	} else if len(signatureList) == 1 {
		return signatureList[0], nil
	} else {
		return "", errors.New("the passed signatureList contains more than one signature value and it is illegal")
	}
}

func (c *l2Client) GenerateSignBody(txData interface{}, ops *types.TransactOpts) (string, error) {
	txInfo, err := c.constructTransaction(txData, ops)
	if err != nil {
		return "", err
	}
	signatureBody := txInfo.GetL1SignatureBody()
	return signatureBody, nil
}

func (c *l2Client) GenerateSignature(privateKey string, txData interface{}) (string, error) {
	l1Signer, err := signer.NewL1Singer(privateKey)
	if err != nil {
		return "", err
	}
	txInfo, err := c.constructTransaction(txData, nil)
	if err != nil {
		return "", err
	}
	signBody := txInfo.GetL1SignatureBody()
	if err != nil {
		return "", err
	}
	signHex, err := l1Signer.Sign(signBody)
	if err != nil {
		return "", err
	}
	return signHex, nil
}

func (c *l2Client) constructTransaction(tx interface{}, ops *types.TransactOpts) (txtypes.TxInfo, error) {
	if ops == nil {
		ops = new(types.TransactOpts)
	}
	if value, ok := tx.(*types.MintNftTxReq); ok {
		return c.constructMintNftTransaction(value, ops)
	} else if value, ok := tx.(*types.CreateCollectionTxReq); ok {
		return c.constructCreateCollectionTransaction(value, ops)
	} else if value, ok := tx.(*types.CancelOfferTxReq); ok {
		return c.constructCancelOfferTransaction(value, ops)
	} else if value, ok := tx.(*types.OfferTxInfo); ok {
		return c.constructOfferTxInfoTransaction(value, ops)
	} else if value, ok := tx.(*types.TransferTxReq); ok {
		return c.constructTransferTransaction(value, ops)
	} else if value, ok := tx.(*types.TransferNftTxReq); ok {
		return c.constructTransferNftTransaction(value, ops)
	} else if value, ok := tx.(*types.WithdrawTxReq); ok {
		return c.constructWithdrawTransaction(value, ops)
	} else if value, ok := tx.(*types.WithdrawNftTxReq); ok {
		return c.constructWithdrawNftTransaction(value, ops)
	} else if value, ok := tx.(*types.UpdateNftReq); ok {
		return c.constructUpdateNFTTransaction(value, ops)
	} else if value, ok := tx.(*types.ChangePubKeyReq); ok {
		return c.constructChangePubKeyTransaction(value, ops)
	}
	return nil, errors.New("invalid tx type is passed")
}

func (c *l2Client) constructChangePubKeyTransaction(tx *types.ChangePubKeyReq, ops *types.TransactOpts) (*txtypes.ChangePubKeyInfo, error) {
	ops, err := c.fullFillChangePubKeyOps(ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := txutils.ConstructChangePubKeyTx(c.keyManager, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *l2Client) constructMintNftTransaction(tx *types.MintNftTxReq, ops *types.TransactOpts) (*txtypes.MintNftTxInfo, error) {
	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}

	ops, err = c.fullFillToAddrOps(ops, tx.To)
	if err != nil {
		return nil, err
	}

	txInfo, err := txutils.ConstructMintNftTx(c.keyManager, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *l2Client) constructCancelOfferTransaction(tx *types.CancelOfferTxReq, ops *types.TransactOpts) (*txtypes.CancelOfferTxInfo, error) {

	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := txutils.ConstructCancelOfferTx(c.keyManager, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *l2Client) constructCreateCollectionTransaction(tx *types.CreateCollectionTxReq, ops *types.TransactOpts) (*txtypes.CreateCollectionTxInfo, error) {
	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := txutils.ConstructCreateCollectionTx(c.keyManager, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *l2Client) constructAtomicMatchTransaction(tx *types.AtomicMatchTxReq, ops *types.TransactOpts) (*txtypes.AtomicMatchTxInfo, error) {

	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := txutils.ConstructAtomicMatchTx(c.keyManager, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *l2Client) constructOfferTxInfoTransaction(tx *types.OfferTxInfo, ops *types.TransactOpts) (*txtypes.OfferTxInfo, error) {
	return &txtypes.OfferTxInfo{
		Type:         tx.Type,
		OfferId:      tx.OfferId,
		AccountIndex: tx.AccountIndex,
		NftIndex:     tx.NftIndex,
		AssetId:      tx.AssetId,
		AssetAmount:  tx.AssetAmount,
		ListedAt:     tx.ListedAt,
		ExpiredAt:    tx.ExpiredAt,
		TreasuryRate: tx.TreasuryRate,
		Sig:          tx.Sig,
	}, nil
}

func (c *l2Client) constructTransferNftTransaction(tx *types.TransferNftTxReq, ops *types.TransactOpts) (*txtypes.TransferNftTxInfo, error) {
	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	ops, err = c.fullFillToAddrOps(ops, tx.To)
	if err != nil {
		return nil, err
	}
	txInfo, err := txutils.ConstructTransferNftTx(c.keyManager, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *l2Client) constructTransferTransaction(tx *types.TransferTxReq, ops *types.TransactOpts) (*txtypes.TransferTxInfo, error) {
	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	ops, err = c.fullFillToAddrOps(ops, tx.To)
	if err != nil {
		return nil, err
	}
	txInfo, err := txutils.ConstructTransferTx(c.keyManager, ops, tx)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *l2Client) constructWithdrawTransaction(tx *types.WithdrawTxReq, ops *types.TransactOpts) (*txtypes.WithdrawTxInfo, error) {

	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := txutils.ConstructWithdrawTxInfo(c.keyManager, tx, ops)
	if err != nil {
		return nil, err
	}

	return txInfo, nil
}

func (c *l2Client) constructWithdrawNftTransaction(tx *types.WithdrawNftTxReq, ops *types.TransactOpts) (*txtypes.WithdrawNftTxInfo, error) {

	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}

	txInfo, err := txutils.ConstructWithdrawNftTx(c.keyManager, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *l2Client) constructUpdateNFTTransaction(req *types.UpdateNftReq, ops *types.TransactOpts) (*txtypes.UpdateNFTTxInfo, error) {
	updateNFTTxInfo, err := txutils.ConstructUpdateNFTTx(req, ops)
	if err != nil {
		return nil, err
	}
	return updateNFTTxInfo, nil
}
