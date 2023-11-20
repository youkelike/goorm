package reflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterateArrayOrSlice(t *testing.T) {
	testCases := []struct {
		name     string
		entity   any
		wantVals any
		wantErr  error
	}{
		{
			name:     "slice",
			entity:   []int{1, 3, 4, 5},
			wantVals: []any{1, 3, 4, 5},
		},
		{
			name:     "array",
			entity:   [4]int{1, 3, 4, 5},
			wantVals: []any{1, 3, 4, 5},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateArrayOrSlice(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVals, res)
		})
	}
}

func TestIterateMap(t *testing.T) {
	testCases := []struct {
		name     string
		entity   any
		wantKeys []any
		wantVals []any
		wantErr  error
	}{
		{
			name: "map",
			entity: map[string]string{
				"A": "a",
				"B": "b",
			},
			wantKeys: []any{"A", "B"},
			wantVals: []any{"a", "b"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys, vals, err := IterateMap(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.EqualValues(t, tc.wantKeys, keys)
			assert.EqualValues(t, tc.wantVals, vals)
		})
	}

}
