package material

import (
	"context"
	"database/sql"
	"errors"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"multi-system/mes/internal/model"
)

type querier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func readRecord(ctx context.Context, q querier, id string, lock bool) (*model.Record, error) {
	suffix := ""
	if lock {
		suffix = " FOR UPDATE"
	}
	r := new(model.Record)
	var date, created, updated time.Time
	err := q.QueryRowContext(ctx, `SELECT id,number,process_id,warehouse_code,requested_date,remark,sync_status,last_error,attempts,created_at,updated_at FROM material_requisition WHERE id=?`+suffix, id).Scan(&r.ID, &r.Number, &r.ProcessID, &r.WarehouseID, &date, &r.Remark, &r.Status, &r.LastError, &r.Attempts, &created, &updated)
	if err != nil {
		return nil, err
	}
	r.ClientRequestID = r.ID
	r.Date = date.Format("2006-01-02")
	r.CreatedAt = created.UTC().Format(time.RFC3339)
	r.UpdatedAt = updated.UTC().Format(time.RFC3339)
	rows, err := q.QueryContext(ctx, `SELECT material_code,requested_qty,remark FROM material_requisition_line WHERE requisition_id=? ORDER BY line_no`, id)
	if err != nil {
		return nil, err
	}
	r.Items = []model.Line{}
	for rows.Next() {
		var line model.Line
		if err = rows.Scan(&line.MaterialCode, &line.Quantity, &line.Remark); err != nil {
			rows.Close()
			return nil, err
		}
		r.Items = append(r.Items, line)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = q.QueryContext(ctx, `SELECT created_at,message FROM integration_event_log WHERE requisition_id=? ORDER BY id DESC LIMIT 30`, id)
	if err != nil {
		return nil, err
	}
	r.Events = []model.Event{}
	for rows.Next() {
		var event model.Event
		var timestamp time.Time
		if err = rows.Scan(&timestamp, &event.Message); err != nil {
			rows.Close()
			return nil, err
		}
		event.Time = timestamp.UTC().Format(time.RFC3339)
		r.Events = append(r.Events, event)
	}
	err = rows.Err()
	rows.Close()
	for i, j := 0, len(r.Events)-1; i < j; i, j = i+1, j-1 {
		r.Events[i], r.Events[j] = r.Events[j], r.Events[i]
	}
	return r, err
}
func (s *Service) Get(ctx context.Context, id string) (*model.Record, error) {
	return readRecord(ctx, s.DB, id, false)
}
func (s *Service) CreateRecord(ctx context.Context, r *model.Record) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO material_requisition (id,number,request_hash,process_id,warehouse_code,requested_date,remark,sync_status,last_error,attempts,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,UTC_TIMESTAMP(6),UTC_TIMESTAMP(6))`, r.ID, r.Number, digest(r.Submission), r.ProcessID, r.WarehouseID, r.Date, r.Remark, r.Status, "", 0)
	if err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			tx.Rollback()
			var previousHash string
			if err = s.DB.QueryRowContext(ctx, `SELECT request_hash FROM material_requisition WHERE id=?`, r.ID).Scan(&previousHash); err != nil {
				return err
			}
			if previousHash != digest(r.Submission) {
				return errors.New("同一请求标识不能提交不同内容")
			}
			return nil
		}
		return err
	}
	for i, line := range r.Items {
		var material model.Material
		for _, m := range model.Materials {
			if m.Code == line.MaterialCode {
				material = m
				break
			}
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO material_requisition_line (requisition_id,line_no,material_code,material_name,spec,unit,requested_qty,remark) VALUES (?,?,?,?,?,?,?,?)`, r.ID, i+1, line.MaterialCode, material.Name, material.Spec, material.Unit, line.Quantity, line.Remark)
		if err != nil {
			return err
		}
	}
	for _, event := range r.Events {
		if _, err = tx.ExecContext(ctx, `INSERT INTO integration_event_log (requisition_id,message) VALUES (?,?)`, r.ID, event.Message); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SetStatus records queue/consumption progress without overwriting a completed task.
func (s *Service) SetStatus(ctx context.Context, id, status, message string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	attempts := 0
	if status == "consumed" {
		attempts = 1
	}
	lastError := ""
	if status == "enqueue_failed" {
		lastError = message
	}
	result, err := tx.ExecContext(ctx, `UPDATE material_requisition SET sync_status=?,last_error=?,attempts=attempts+?,updated_at=UTC_TIMESTAMP(6) WHERE id=? AND sync_status<>'consumed'`, status, lastError, attempts, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count > 0 {
		if _, err = tx.ExecContext(ctx, `INSERT INTO integration_event_log (requisition_id,message) VALUES (?,?)`, id, message); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (s *Service) List(ctx context.Context) ([]model.Record, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id FROM material_requisition ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := make([]model.Record, 0, len(ids))
	for _, id := range ids {
		r, e := s.Get(ctx, id)
		if e != nil {
			return nil, e
		}
		result = append(result, *r)
	}
	return result, nil
}
