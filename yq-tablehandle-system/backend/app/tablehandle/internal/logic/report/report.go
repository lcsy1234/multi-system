package report

import (
	"context"
	"github.com/xuri/excelize/v2"
)

func Sample() (*excelize.File, error) { return sample(today()) }
func Options(mode string) map[string]any {
	now := today()
	return map[string]any{"factories": factories, "lines": lines, "skus": skus, "start": date(now.AddDate(0, 0, -13)), "end": date(now), "mode": mode}
}
func Run(ctx context.Context, book *excelize.File, sheet, id, dir string, filters Filters, trace func(string, string)) (Report, error) {
	job := Job{book: book, sheet: sheet, id: id, dir: dir, filters: filters, db: seed(today()), trace: trace}
	err := job.run(ctx)
	return job.report, err
}
