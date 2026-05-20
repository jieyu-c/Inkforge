package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ConsolePromptsModel = (*customConsolePromptsModel)(nil)

type (
	// ConsolePromptsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConsolePromptsModel.
	ConsolePromptsModel interface {
		consolePromptsModel
		withSession(session sqlx.Session) ConsolePromptsModel
		CountByTenantNs(ctx context.Context, tenantID, nsID, qPrefix string) (int64, error)
		ListByTenantNs(ctx context.Context, tenantID, nsID, qPrefix string, offset, limit int) ([]*ConsolePrompts, error)
	}

	customConsolePromptsModel struct {
		*defaultConsolePromptsModel
	}
)

// NewConsolePromptsModel returns a model for the database table.
func NewConsolePromptsModel(conn sqlx.SqlConn) ConsolePromptsModel {
	return &customConsolePromptsModel{
		defaultConsolePromptsModel: newConsolePromptsModel(conn),
	}
}

func (m *customConsolePromptsModel) withSession(session sqlx.Session) ConsolePromptsModel {
	return NewConsolePromptsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customConsolePromptsModel) CountByTenantNs(ctx context.Context, tenantID, nsID, qPrefix string) (int64, error) {
	where := "`tenant_id` = ? and `ns_id` = ?"
	args := []any{tenantID, nsID}
	if qPrefix != "" {
		where += " and `prompt_key` like ?"
		args = append(args, qPrefix+`%`)
	}
	query := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	var n int64
	err := m.conn.QueryRowCtx(ctx, &n, query, args...)
	return n, err
}

func (m *customConsolePromptsModel) ListByTenantNs(ctx context.Context, tenantID, nsID, qPrefix string, offset, limit int,
) ([]*ConsolePrompts, error) {
	where := "`tenant_id` = ? and `ns_id` = ?"
	args := []any{tenantID, nsID}
	if qPrefix != "" {
		where += " and `prompt_key` like ?"
		args = append(args, qPrefix+`%`)
	}
	query := fmt.Sprintf("select %s from %s where %s order by `prompt_key` asc limit ? offset ?", consolePromptsRows, m.table, where)
	args = append(args, limit, offset)
	var rows []ConsolePrompts
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	out := make([]*ConsolePrompts, 0, len(rows))
	for i := range rows {
		r := rows[i]
		cp := r
		out = append(out, &cp)
	}
	return out, nil
}
