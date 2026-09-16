// Package audit records before/after snapshots of mutated rows into
// audit_logs. Call Log from within the same transaction as the mutation it
// documents so the audit trail and the change it describes commit or roll
// back together.
package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ranji/clothing-erp/internal/db"
)

type Action string

const (
	Insert Action = "INSERT"
	Update Action = "UPDATE"
	Delete Action = "DELETE"
)

func Log(ctx context.Context, q db.Querier, tableName string, recordID uuid.UUID, action Action, oldData, newData any) error {
	var oldJSON, newJSON []byte
	var err error

	if oldData != nil {
		if oldJSON, err = json.Marshal(oldData); err != nil {
			return err
		}
	}
	if newData != nil {
		if newJSON, err = json.Marshal(newData); err != nil {
			return err
		}
	}

	_, err = q.Exec(ctx, `
		INSERT INTO audit_logs (table_name, record_id, action, old_data, new_data, changed_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tableName, recordID, string(action), oldJSON, newJSON, db.ActorFromContext(ctx))
	return err
}
