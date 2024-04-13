package sentinel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Sentinel struct {
	APIEndpoint string
	RPCEndpoint string

	ProviderPlanBlockchainID string

	FeeGranterWalletAddress string
	FeeGranterMnemonic      string

	DefaultDenom string
	ChainID      string
	GasPrice     string
	GasBase      int64
}

func (s Sentinel) FetchFeeGrantAllowances(walletAddress string, limit int, offset int) (*[]SentinelAllowance, error) {
	type blockchainResponse struct {
		Success bool                 `json:"success"`
		Error   *SentinelError       `json:"error"`
		Result  *[]SentinelAllowance `json:"result"`
	}

	args := fmt.Sprintf(
		"?rpc_address=%s&chain_id=%s&limit=%d&offset=%d",
		s.RPCEndpoint,
		s.ChainID,
		limit,
		offset,
	)

	url := s.APIEndpoint + "/api/v1/feegrants/" + walletAddress + "/allowances" + args
	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response *blockchainResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Success == false {
		apiError := ""
		if response.Error != nil {
			apiError = " (" + response.Error.Message + ")"
		}

		return nil, errors.New("success `false` returned from Sentinel API when fetching fee grant allowances: " + apiError)
	}

	return response.Result, nil
}

func (s Sentinel) GrantFeeToWallet(walletAddresses []string) error {
	type blockchainResponse struct {
		Success bool                 `json:"success"`
		Error   *SentinelError       `json:"error"`
		Result  *SentinelTransaction `json:"result"`
	}

	type blockchainRequest struct {
		Mnemonic     string   `json:"mnemonic"`
		AccAddresses []string `json:"acc_addresses"`
		AllowedMsgs  []string `json:"allowed_msgs"`
	}

	payload, err := json.Marshal(blockchainRequest{
		Mnemonic:     s.FeeGranterMnemonic,
		AccAddresses: walletAddresses,
		AllowedMsgs:  []string{"/sentinel.plan.v2.MsgSubscribeRequest", "/sentinel.session.v2.MsgStartRequest", "/sentinel.session.v2.MsgEndRequest"},
	})

	if err != nil {
		return err
	}

	gas := s.GasBase * int64(len(walletAddresses)+1)

	args := fmt.Sprintf(
		"?rpc_address=%s&chain_id=%s&gas_prices=%s&gas=%d&simulate_and_execute=false",
		s.RPCEndpoint,
		s.ChainID,
		s.GasPrice+s.DefaultDenom,
		gas,
	)

	url := s.APIEndpoint + "/api/v1/feegrants" + args
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var response *blockchainResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	if response.Success == false {
		apiError := ""
		if response.Error != nil {
			apiError = " (" + response.Error.Message + ")"
		}

		return errors.New("success `false` returned from Sentinel API while granting fee to wallets" + apiError)
	}

	return nil
}
