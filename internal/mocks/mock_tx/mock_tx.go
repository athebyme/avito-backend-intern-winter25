package mocks_tx

import (
	_ "database/sql"
	"github.com/stretchr/testify/mock"
)

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit() error {
	return m.Called().Error(0)
}

func (m *MockTx) Rollback() error {
	return m.Called().Error(0)
}
