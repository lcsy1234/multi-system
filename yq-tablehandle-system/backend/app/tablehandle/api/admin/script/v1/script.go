package v1

const OptimizePath = "/script-optimizations"

type OptimizeReq struct{ Sample, Instruction, Actual, Previous string }
type OptimizeRes struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
