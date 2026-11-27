package tablehandle

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/net/ghttp"
	"multi-system/yq-tablehandle-system/backend/app/tablehandle/api/admin/conversion"
	cv1 "multi-system/yq-tablehandle-system/backend/app/tablehandle/api/admin/conversion/v1"
	"multi-system/yq-tablehandle-system/backend/app/tablehandle/api/admin/script"
	sv1 "multi-system/yq-tablehandle-system/backend/app/tablehandle/api/admin/script/v1"
	"net/http"
)

type conversionController struct{}
type scriptController struct{}
type apiError struct {
	status  int
	message string
}

func (e apiError) Error() string { return e.message }
func bindAdmin(group *ghttp.RouterGroup, conversions conversion.IConversionV1, scripts script.IScriptV1) {
	const invalidContent = "请输入非空文本，最大 64 KiB"
	group.POST(cv1.ConvertPath, jsonHandler(6*conversionLimit+128, invalidContent, conversions.Convert))
	group.POST(cv1.ConvertShrimpPath, jsonHandler(6*conversionLimit+128, invalidContent, conversions.ConvertShrimp))
	group.POST(sv1.OptimizePath, jsonHandler(1<<20, "请填写样本（最多 64 KiB）和处理要求（最多 8 KiB）", scripts.Optimize))
}
func jsonHandler[Req, Res any](limit int64, invalid string, handle func(context.Context, *Req) (*Res, error)) ghttp.HandlerFunc {
	return func(c *ghttp.Request) {
		var input Req
		c.Body = http.MaxBytesReader(c.Response.RawWriter(), c.Body, limit)
		if json.NewDecoder(c.Body).Decode(&input) != nil {
			writeJSON(c, 400, map[string]string{"error": invalid})
			return
		}
		result, err := handle(c.Context(), &input)
		if err != nil {
			status := 500
			if failure, ok := err.(apiError); ok {
				status = failure.status
			}
			writeJSON(c, status, map[string]string{"error": err.Error()})
			return
		}
		if raw, ok := any(result).(*json.RawMessage); ok {
			c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
			c.Response.Write([]byte(*raw))
			return
		}
		writeJSON(c, 200, result)
	}
}
