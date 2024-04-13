package sentinel

import "time"

type SentinelError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type SentinelTransactionEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SentinelTransactionEvent struct {
	Type       string                              `json:"type"`
	Attributes []SentinelTransactionEventAttribute `json:"attributes"`
}

type SentinelTransactionResult struct {
	Events []SentinelTransactionEvent `json:"events"`
}

type SentinelTransaction struct {
	Height   int64                     `json:"height"`
	TxHash   string                    `json:"txhash"`
	TxResult SentinelTransactionResult `json:"tx_result"`
}

type SentinelAllowanceDetails struct {
	Expiration *time.Time `json:"expiration"`
}

type SentinelAllowance struct {
	Allowance SentinelAllowanceDetails `json:"allowance"`
	Grantee   string                   `json:"grantee"`
	Granter   string                   `json:"granter"`
}
