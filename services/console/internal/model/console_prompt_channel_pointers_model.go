package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ConsolePromptChannelPointersModel = (*customConsolePromptChannelPointersModel)(nil)

type (
	// ConsolePromptChannelPointersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConsolePromptChannelPointersModel.
	ConsolePromptChannelPointersModel interface {
		consolePromptChannelPointersModel
		withSession(session sqlx.Session) ConsolePromptChannelPointersModel
	}

	customConsolePromptChannelPointersModel struct {
		*defaultConsolePromptChannelPointersModel
	}
)

// NewConsolePromptChannelPointersModel returns a model for the database table.
func NewConsolePromptChannelPointersModel(conn sqlx.SqlConn) ConsolePromptChannelPointersModel {
	return &customConsolePromptChannelPointersModel{
		defaultConsolePromptChannelPointersModel: newConsolePromptChannelPointersModel(conn),
	}
}

func (m *customConsolePromptChannelPointersModel) withSession(session sqlx.Session) ConsolePromptChannelPointersModel {
	return NewConsolePromptChannelPointersModel(sqlx.NewSqlConnFromSession(session))
}
