package v1

import "encoding/json"

const ConvertPath = "/json-conversions"
const ConvertShrimpPath = "/shrimp/json-conversions"

type ConvertReq struct {
	Content string `json:"content"`
}
type ConvertRes = json.RawMessage
type ConvertShrimpReq = ConvertReq
type ConvertShrimpRes = ConvertRes
