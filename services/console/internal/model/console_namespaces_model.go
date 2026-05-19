package model

import (
	"context"
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
