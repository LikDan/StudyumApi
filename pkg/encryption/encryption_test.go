package encryption

import (
	"github.com/go-playground/assert/v2"
	"reflect"
	"testing"
)

func TestEncryption_MapReflectField(t *testing.T) {
	type s struct {
		Int          int    `encryption:""`
		String       string `encryption:""`
		String2      string
		StringArray  []string `encryption:""`
		StringArray2 []string ``
		IntArray     []int    `encryption:""`
		CustomStruct struct {
			String string `encryption:""`
		} `encryption:""`
	}

	type args struct {
		value s
	}
	tests := []struct {
		name string
		args args
		want s
	}{
		{
			name: "Valid test",
			args: args{
				value: s{
					Int:          10,
					String:       "a",
					String2:      "a",
					StringArray:  []string{"1", "2", "3"},
					StringArray2: []string{"1", "2", "3"},
					IntArray:     []int{1, 2, 3},
					CustomStruct: struct {
						String string `encryption:""`
					}{"a"},
				},
			},
			want: s{
				Int:          10,
				String:       "encrypted",
				String2:      "a",
				StringArray:  []string{"encrypted", "encrypted", "encrypted"},
				StringArray2: []string{"1", "2", "3"},
				IntArray:     []int{1, 2, 3},
				CustomStruct: struct {
					String string `encryption:""`
				}{"encrypted"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &encryption{}
			e.MapReflectField(reflect.ValueOf(&tt.args.value).Elem(), func(str string) string {
				return "encrypted"
			})

			assert.Equal(t, tt.args.value, tt.want)
		})
	}
}
