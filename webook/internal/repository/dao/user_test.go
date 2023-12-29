package dao

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string

		//这里是sqlmock，不是gomock
		mock func(t *testing.T) *sql.DB

		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()

				res := sqlmock.NewResult(3, 1)
				//这边预期的是正则表达式
				//这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
			user: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      tc.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			d := NewUserDAO(db)
			err = d.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)

		})

	}
}
