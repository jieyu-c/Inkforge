package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ConsolePromptVersionsModel = (*customConsolePromptVersionsModel)(nil)

type (
	// ConsolePromptVersionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConsolePromptVersionsModel.
	ConsolePromptVersionsModel interface {
		consolePromptVersionsModel
		withSession(session sqlx.Session) ConsolePromptVersionsModel
		CountByPromptScoped(ctx context.Context, tenantID, nsID, promptID, versionQ string) (int64, error)
		MaxVersionNum(ctx context.Context, tenantID, nsID, promptID string) (uint64, error)
		ListByPromptScoped(ctx context.Context, tenantID, nsID, promptID, versionQ string, offset, limit int) ([]*ConsolePromptVersions, error)
		FindOneByPromptVersionNum(ctx context.Context, tenantID, nsID, promptID string, versionNum uint64) (*ConsolePromptVersions, error)
		FindOneByPromptVersionLabel(ctx context.Context, tenantID, nsID, promptID, label string) (*ConsolePromptVersions, error)
		ListVersionLabelsByPrompt(ctx context.Context, tenantID, nsID, promptID string) ([]string, error)
	}

	customConsolePromptVersionsModel struct {
		*defaultConsolePromptVersionsModel
	}
)

// NewConsolePromptVersionsModel returns a model for the database table.
func NewConsolePromptVersionsModel(conn sqlx.SqlConn) ConsolePromptVersionsModel {
	return &customConsolePromptVersionsModel{
		defaultConsolePromptVersionsModel: newConsolePromptVersionsModel(conn),
	}
}

func (m *customConsolePromptVersionsModel) withSession(session sqlx.Session) ConsolePromptVersionsModel {
	return NewConsolePromptVersionsModel(sqlx.NewSqlConnFromSession(session))
}

func versionSearchClause(versionQ string) (string, []any) {
	raw := strings.TrimSpace(versionQ)
	raw = strings.TrimPrefix(strings.ToLower(raw), "v")
	if raw == "" {
		return "", nil
	}
	return " and `version_label` like ?", []any{raw + "%"}
}

func (m *customConsolePromptVersionsModel) CountByPromptScoped(ctx context.Context, tenantID, nsID, promptID, versionQ string) (int64, error) {
	extra, args := versionSearchClause(versionQ)
	q := fmt.Sprintf("select count(*) from %s where `tenant_id` = ? and `ns_id` = ? and `prompt_id` = ?%s", m.table, extra)
	params := append([]any{tenantID, nsID, promptID}, args...)
	var n int64
	err := m.conn.QueryRowCtx(ctx, &n, q, params...)
	return n, err
}

func (m *customConsolePromptVersionsModel) MaxVersionNum(ctx context.Context, tenantID, nsID, promptID string) (uint64, error) {
	q := fmt.Sprintf("select coalesce(max(`version_num`), 0) from %s where `tenant_id` = ? and `ns_id` = ? and `prompt_id` = ?", m.table)
	var n uint64
	err := m.conn.QueryRowCtx(ctx, &n, q, tenantID, nsID, promptID)
	return n, err
}

func (m *customConsolePromptVersionsModel) FindOneByPromptVersionLabel(
	ctx context.Context, tenantID, nsID, promptID, label string,
) (*ConsolePromptVersions, error) {
	q := fmt.Sprintf("select %s from %s where `tenant_id` = ? and `ns_id` = ? and `prompt_id` = ? and `version_label` = ? limit 1",
		consolePromptVersionsRows, m.table)
	var row ConsolePromptVersions
	err := m.conn.QueryRowCtx(ctx, &row, q, tenantID, nsID, promptID, label)
	switch err {
	case nil:
		return &row, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customConsolePromptVersionsModel) ListVersionLabelsByPrompt(
	ctx context.Context, tenantID, nsID, promptID string,
) ([]string, error) {
	q := fmt.Sprintf("select `version_label` from %s where `tenant_id` = ? and `ns_id` = ? and `prompt_id` = ?",
		m.table)
	var rows []struct {
		VersionLabel string `db:"version_label"`
	}
	if err := m.conn.QueryRowsCtx(ctx, &rows, q, tenantID, nsID, promptID); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.VersionLabel != "" {
			out = append(out, r.VersionLabel)
		}
	}
	return out, nil
}

func (m *customConsolePromptVersionsModel) FindOneByPromptVersionNum(
	ctx context.Context, tenantID, nsID, promptID string, versionNum uint64,
) (*ConsolePromptVersions, error) {
	q := fmt.Sprintf("select %s from %s where `tenant_id` = ? and `ns_id` = ? and `prompt_id` = ? and `version_num` = ? limit 1",
		consolePromptVersionsRows, m.table)
	var row ConsolePromptVersions
	err := m.conn.QueryRowCtx(ctx, &row, q, tenantID, nsID, promptID, versionNum)
	switch err {
	case nil:
		return &row, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customConsolePromptVersionsModel) ListByPromptScoped(ctx context.Context, tenantID, nsID, promptID, versionQ string, offset, limit int,
) ([]*ConsolePromptVersions, error) {
	extra, args := versionSearchClause(versionQ)
	q := fmt.Sprintf("select %s from %s where `tenant_id` = ? and `ns_id` = ? and `prompt_id` = ?%s order by `version_num` desc limit ? offset ?",
		consolePromptVersionsRows, m.table, extra)
	params := append([]any{tenantID, nsID, promptID}, args...)
	params = append(params, limit, offset)
	var rows []ConsolePromptVersions
	if err := m.conn.QueryRowsCtx(ctx, &rows, q, params...); err != nil {
		return nil, err
	}
	out := make([]*ConsolePromptVersions, 0, len(rows))
	for i := range rows {
		r := rows[i]
		cp := r
		out = append(out, &cp)
	}
	return out, nil
}
