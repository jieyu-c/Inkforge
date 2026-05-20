package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ConsoleNamespacesModel = (*customConsoleNamespacesModel)(nil)

type (
	// ConsoleNamespacesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConsoleNamespacesModel.
	ConsoleNamespacesModel interface {
		consoleNamespacesModel
		withSession(session sqlx.Session) ConsoleNamespacesModel
		ListByTenantId(ctx context.Context, tenantID string) ([]*ConsoleNamespaces, error)
		// AddPromptCount increments console_namespaces.prompt_count when a new prompt_key is created (delta must be +1 or -1).
		AddPromptCount(ctx context.Context, tenantID, nsID string, delta int) error
	}

	customConsoleNamespacesModel struct {
		*defaultConsoleNamespacesModel
	}
)

// NewConsoleNamespacesModel returns a model for the database table.
func NewConsoleNamespacesModel(conn sqlx.SqlConn) ConsoleNamespacesModel {
	return &customConsoleNamespacesModel{
		defaultConsoleNamespacesModel: newConsoleNamespacesModel(conn),
	}
}

func (m *customConsoleNamespacesModel) withSession(session sqlx.Session) ConsoleNamespacesModel {
	return NewConsoleNamespacesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customConsoleNamespacesModel) AddPromptCount(ctx context.Context, tenantID, nsID string, delta int) error {
	if delta == 0 {
		return nil
	}
	var q string
	switch {
	case delta > 0:
		q = fmt.Sprintf(`update %s set prompt_count = prompt_count + ? where id = ? and tenant_id = ?`, "`console_namespaces`")
	default:
		q = fmt.Sprintf(`update %s set prompt_count = IF(prompt_count > 0, prompt_count - 1, 0) where id = ? and tenant_id = ?`, "`console_namespaces`")
	}
	var res sql.Result
	var err error
	if delta > 0 {
		res, err = m.conn.ExecCtx(ctx, q, delta, nsID, tenantID)
	} else {
		res, err = m.conn.ExecCtx(ctx, q, nsID, tenantID)
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrNotFound
	}
	return nil
}

func (m *customConsoleNamespacesModel) ListByTenantId(ctx context.Context,
	tenantID string,
) ([]*ConsoleNamespaces, error) {
	query := fmt.Sprintf(`select %s from %s where tenant_id = ? order by created_at asc, ns_slug asc`,
		consoleNamespacesRows, "`console_namespaces`")
	var rows []ConsoleNamespaces
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, tenantID); err != nil {
		return nil, err
	}
	out := make([]*ConsoleNamespaces, 0, len(rows))
	for i := range rows {
		r := rows[i]
		copy := r
		out = append(out, &copy)
	}
	return out, nil
}

// Exported for tests (covers tenant_id scoping in SQL).
func ListNamespacesSqlShape() string {
	return strings.TrimSpace(fmt.Sprintf(`select %s from %s where tenant_id = ? order by`,
		consoleNamespacesRows, "`console_namespaces`"))
}
