package valuer

import (
	"database/sql"
	"testing"

	"gitee.com/youkelike/orm/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReflect_SetColumns(t *testing.T) {
	testSetColumns(t, NewReflectValue)
}
func TestUnsafe_SetColumns(t *testing.T) {
	testSetColumns(t, NewUnsafeValue)
}

func testSetColumns(t *testing.T, creator Creator) {
	testCases := []struct {
		name       string
		entity     any
		rows       *sqlmock.Rows
		wantErr    error
		wantEntity any
	}{
		{
			name:   "set columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow("1", "Tom", "18", "Jerry")
				return rows
			}(),
			wantEntity: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       int8(18),
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			},
		},
		{
			name:   "disorder columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"first_name", "age", "id", "last_name"})
				rows.AddRow("Tom", "18", "1", "Jerry")
				return rows
			}(),
			wantEntity: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       int8(18),
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			},
		},
		{
			name:   "partial columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "last_name"})
				rows.AddRow("1", "Jerry")
				return rows
			}(),
			wantEntity: &TestModel{
				Id:       1,
				LastName: &sql.NullString{String: "Jerry", Valid: true},
			},
		},
	}

	r := model.NewRegistry()

	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRows := tc.rows
			mock.ExpectQuery("SELECT XXX").WillReturnRows(mockRows)
			rows, err := mockDB.Query("SELECT XXX")
			require.NoError(t, err)

			rows.Next()

			model, err := r.Get(tc.entity)
			require.NoError(t, err)
			if err != nil {
				return
			}
			val := creator(model, tc.entity)
			err = val.SetColumns(rows)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantEntity, tc.entity)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
