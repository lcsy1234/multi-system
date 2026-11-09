package v1

import "github.com/gogf/gf/v2/frame/g"

// HealthReq only checks that this process can serve requests.
type HealthReq struct {
	g.Meta `path:"/health" method:"get" tags:"Health" summary:"Process liveness"`
}

type HealthRes struct {
	Service string `json:"service" dc:"Service name"`
	Status  string `json:"status" dc:"Process status"`
	Role    string `json:"role" dc:"Service responsibility"`
}
