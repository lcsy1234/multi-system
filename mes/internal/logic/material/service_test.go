package material

import (
	"multi-system/mes/internal/model"
	"testing"
)

func TestValidateSubmission(t *testing.T) {
	valid := func() model.Submission {
		return model.Submission{ClientRequestID: "e2edf4db-98be-47d0-b2f7-d152b28f94b5", ProcessID: "PR20261109001", WarehouseID: "WH001", Date: "2026-11-09", Items: []model.Line{{MaterialCode: "RM001", Quantity: "0.0001"}}}
	}
	tests := []struct {
		name   string
		change func(*model.Submission)
	}{
		{"zero", func(s *model.Submission) { s.Items[0].Quantity = "0" }},
		{"negative", func(s *model.Submission) { s.Items[0].Quantity = "-1" }},
		{"excess precision", func(s *model.Submission) { s.Items[0].Quantity = "0.00001" }},
		{"exponent", func(s *model.Submission) { s.Items[0].Quantity = "1e4" }},
		{"unknown material", func(s *model.Submission) { s.Items[0].MaterialCode = "missing" }},
		{"duplicate material", func(s *model.Submission) { s.Items = append(s.Items, s.Items[0]) }},
		{"completed process", func(s *model.Submission) { s.ProcessID = "PR20261107004" }},
		{"unknown warehouse", func(s *model.Submission) { s.WarehouseID = "missing" }},
		{"invalid date", func(s *model.Submission) { s.Date = "2026-02-30" }},
		{"empty lines", func(s *model.Submission) { s.Items = nil }},
		{"invalid id", func(s *model.Submission) { s.ClientRequestID = "missing" }},
	}
	if err := Validate(valid()); err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := valid()
			tt.change(&s)
			if Validate(s) == nil {
				t.Fatal("expected validation failure")
			}
		})
	}
}
