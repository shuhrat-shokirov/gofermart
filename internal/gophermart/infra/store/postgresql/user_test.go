//nolint:wrapcheck,gocritic,nolintlint,errcheck,forcetypeassert
package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gofermart/internal/gophermart/core/repositories"
)

func TestPostgresql_CreateUser(t *testing.T) {
	t.Run("successful user creation", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
		mockTx.On("Exec", mock.Anything, "INSERT INTO users (login, password) VALUES ($1, $2);",
			[]interface{}{"login", "password"}).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
		mockTx.On("Exec", mock.Anything, "INSERT INTO balance (login) VALUES ($1);",
			[]interface{}{"login"}).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}

		err := postgres.CreateUser(context.TODO(), "login", "password")

		assert.NoError(t, err)

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("error on begin transaction", func(t *testing.T) {
		mockPool := new(MockPool)

		mockPool.On("Begin", mock.Anything).
			Return((*MockTx)(nil), errors.New("begin transaction error"))

		postgres := &Postgresql{pool: mockPool}

		err := postgres.CreateUser(context.TODO(), "testuser", "testpassword")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can't begin transaction")

		mockPool.AssertExpectations(t)
	})

	t.Run("error on insert user", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
		mockTx.On("Exec", mock.Anything, "INSERT INTO users (login, password) VALUES ($1, $2);",
			mock.MatchedBy(func(args []interface{}) bool {
				return len(args) == 2 && args[0] == "testuser" && args[1] == "testpassword"
			})).Return(pgconn.NewCommandTag(""), errors.New("insert user error"))

		mockTx.On("Rollback", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}

		err := postgres.CreateUser(context.TODO(), "testuser", "testpassword")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can't create user")

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("error on insert balance", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)

		mockTx.On("Exec", mock.Anything, "INSERT INTO users (login, password) VALUES ($1, $2);",
			mock.MatchedBy(func(args []interface{}) bool {
				return len(args) == 2 && args[0] == "testuser" && args[1] == "testpassword"
			})).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)

		mockTx.On("Exec", mock.Anything, "INSERT INTO balance (login) VALUES ($1);",
			mock.MatchedBy(func(args []interface{}) bool {
				return len(args) == 1 && args[0] == "testuser"
			})).Return(pgconn.NewCommandTag(""), errors.New("insert balance error"))

		mockTx.On("Rollback", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}

		err := postgres.CreateUser(context.TODO(), "testuser", "testpassword")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can't add balance")

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

func TestPostgresql_GetUserPassword(t *testing.T) {
	t.Run("successful get user password", func(t *testing.T) {
		mockPool := new(MockPool)
		mockRow := new(MockRow)

		mockPool.On("QueryRow", mock.Anything, "SELECT password FROM users WHERE login = $1;",
			mock.MatchedBy(func(args []interface{}) bool {
				return len(args) == 1 && args[0] == "testuser"
			})).Return(mockRow)

		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*string)) = "hashedpassword"
		}).Return(nil)

		postgres := &Postgresql{pool: mockPool}

		password, err := postgres.GetUserPassword(context.TODO(), "testuser")

		assert.NoError(t, err)
		assert.Equal(t, "hashedpassword", password)

		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("error on scan", func(t *testing.T) {
		mockPool := new(MockPool)
		mockRow := new(MockRow)

		mockPool.On("QueryRow", mock.Anything, "SELECT password FROM users WHERE login = $1;",
			mock.MatchedBy(func(args []interface{}) bool {
				return len(args) == 1 && args[0] == "testuser"
			})).Return(mockRow)

		mockRow.On("Scan", mock.Anything).Return(errors.New("scan error"))

		postgres := &Postgresql{pool: mockPool}

		password, err := postgres.GetUserPassword(context.TODO(), "testuser")

		assert.Error(t, err)
		assert.Empty(t, password)
		assert.Contains(t, err.Error(), "can't scan")

		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("err no rows", func(t *testing.T) {
		mockPool := new(MockPool)
		mockRow := new(MockRow)

		mockPool.On("QueryRow", mock.Anything, "SELECT password FROM users WHERE login = $1;",
			mock.MatchedBy(func(args []interface{}) bool {
				return len(args) == 1 && args[0] == "testuser"
			})).Return(mockRow)

		mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		postgres := &Postgresql{pool: mockPool}

		password, err := postgres.GetUserPassword(context.TODO(), "testuser")

		assert.Error(t, err)
		assert.Empty(t, password)
		assert.ErrorIs(t, err, repositories.ErrNotFound)

		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
