package reflect

import (
	"reflect"
	"testing"

	"gitee.com/youkelike/go1/work/hw04/orm/reflect/types"
	"github.com/stretchr/testify/assert"
)

func TestIterateFunc(t *testing.T) {
	testCases := []struct {
		name    string
		entity  any
		wantRes map[string]FuncInfo
		wantErr error
	}{
		{
			name:   "struct",
			entity: types.NewUser("Tom", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": {
					Name:        "GetAge",
					InputTypes:  []reflect.Type{reflect.TypeOf(types.User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
				},
				// "ChangeName": {
				// 	Name:       "ChangeName",
				// 	InputTypes: []reflect.Type{reflect.TypeOf("")},
				// },
			},
		},
		{
			name:   "pointer",
			entity: types.NewUserPtr("Tom", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": {
					Name: "GetAge",
					// 注意这里，第一个参数传入的也是 *T
					InputTypes:  []reflect.Type{reflect.TypeOf(&types.User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
				},
				"ChangeName": {
					Name: "ChangeName",
					// 第一个参数传入的是 *T
					InputTypes:  []reflect.Type{reflect.TypeOf(&types.User{}), reflect.TypeOf("")},
					OutputTypes: []reflect.Type{},
					Result:      []any{},
				},
				// "ChangeName": {
				// 	Name:       "ChangeName",
				// 	InputTypes: []reflect.Type{reflect.TypeOf("")},
				// },
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateFunc(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
