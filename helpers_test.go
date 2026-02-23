package playwright

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testOptionsJSONSerialization struct {
	StringPointer  *string `json:"stringPointer"`
	NormalString   string  `json:"normalString"`
	WithoutJSONTag string
	WithJSONTag    string   `json:"withJSONTag"`
	OverrideMe     []string `json:"overrideMe"`
	SkipNilPtrs    *string  `json:"skipNilPtrs"`
	SkipMe         *int     `json:"skipMe"`
}

func TestTransformOptions(t *testing.T) {
	// test data
	structVar := &testOptionsJSONSerialization{
		StringPointer:  String("1"),
		NormalString:   "2",
		WithoutJSONTag: "3",
		WithJSONTag:    "4",
		OverrideMe:     []string{"5"},
	}
	var nilStrPtr *string
	testCases := []struct {
		name         string
		firstOption  any
		secondOption any
		expected     any
	}{
		{
			name: "No options supplied",
			firstOption: map[string]any{
				"1234": nilStrPtr,
				"foo":  "bar",
			},
			expected: map[string]any{
				"foo": "bar",
			},
		},
		{
			name: "Options are nil",
			firstOption: map[string]any{
				"foo": "bar",
			},
			secondOption: nil,
			expected: map[string]any{
				"foo": "bar",
			},
		},
		{
			name: "JSON serialization works",
			firstOption: map[string]any{
				"foo":           "bar",
				"stringPointer": 1,
			},
			secondOption: structVar,
			expected: map[string]any{
				"foo":            "bar",
				"stringPointer":  String("1"),
				"normalString":   "2",
				"WithoutJSONTag": "3",
				"withJSONTag":    "4",
				"overrideMe":     []any{"5"},
			},
		},
		{
			name: "Second overwrites the first one",
			firstOption: map[string]any{
				"foo": "1",
			},
			secondOption: map[string]any{
				"foo": "2",
			},
			expected: map[string]any{
				"foo": "2",
			},
		},
		{
			name:        "Second overwrites the first one's value in different type",
			firstOption: structVar,
			secondOption: map[string]any{
				"overrideMe": "5",
			},
			expected: map[string]any{
				"stringPointer":  String("1"),
				"normalString":   "2",
				"WithoutJSONTag": "3",
				"withJSONTag":    "4",
				"overrideMe":     "5",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, transformOptions(tc.firstOption, tc.secondOption))
		})
	}
}

func TestRemapMapToStruct(t *testing.T) {
	ourStruct := struct {
		V1 string `json:"v1"`
	}{}
	inMap := map[string]any{
		"v1": "foobar",
	}
	remapMapToStruct(inMap, &ourStruct)
	require.Equal(t, ourStruct.V1, "foobar")
}

func TestRemapMapToStructShouldKeepNilPointerForMissingField(t *testing.T) {
	ourStruct := struct {
		V1           string  `json:"v1"`
		PartitionKey *string `json:"partitionKey"`
	}{}
	inMap := map[string]any{
		"v1": "foobar",
	}

	remapMapToStruct(inMap, &ourStruct)

	require.Equal(t, "foobar", ourStruct.V1)
	require.Nil(t, ourStruct.PartitionKey)
}

func TestConvertSelectOptionSet(t *testing.T) {
	testCases := []struct {
		name         string
		optionValues SelectOptionValues
		expected     any
	}{
		{
			name:         "SelectOptionValues is nil",
			optionValues: SelectOptionValues{},
			expected:     make(map[string]any),
		},
		{
			name: "SelectOptionValues is supplied",
			optionValues: SelectOptionValues{
				ValuesOrLabels: StringSlice("c", "d"),
				Values:         StringSlice("a", "b"),
				Indexes:        IntSlice(1),
				Labels:         StringSlice("x"),
			},
			expected: map[string]any{
				"options": []map[string]any{
					{"valueOrLabel": "c"}, {"valueOrLabel": "d"}, {"value": "a"}, {"value": "b"}, {"index": 1}, {"label": "x"},
				},
			},
		},
		{
			name: "Only value is supplied",
			optionValues: SelectOptionValues{
				Values: StringSlice("a", "b"),
			},
			expected: map[string]any{
				"options": []map[string]any{
					{"value": "a"}, {"value": "b"},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, convertSelectOptionSet(tc.optionValues))
		})
	}
}

func TestAssignFields(t *testing.T) {
	type (
		A struct {
			Field1 string
			Field2 int
		}
		B struct {
			Field1 string
			Field2 int
			Field3 float64
		}
		args struct {
			dest      any
			src       any
			omitExtra bool
		}
	)
	testV := "foo"

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "src is nil",
			args: args{
				dest:      &B{},
				src:       nil,
				omitExtra: true,
			},
			wantErr: true,
		},
		{
			name: "src is not struct",
			args: args{
				dest:      &B{},
				src:       "foo",
				omitExtra: true,
			},
			wantErr: true,
		},
		{
			name: "dest is nil",
			args: args{
				dest:      nil,
				src:       &A{},
				omitExtra: true,
			},
			wantErr: true,
		},
		{
			name: "dest is not struct",
			args: args{
				dest:      &testV,
				src:       &A{},
				omitExtra: true,
			},
			wantErr: true,
		},
		{
			name: "dest includes all src fields",
			args: args{
				dest: &B{},
				src: &A{
					Field1: "hello",
					Field2: 42,
				},
				omitExtra: true,
			},
			wantErr: false,
		},
		{
			name: "dest does not include all src fields, omit extra fields",
			args: args{
				dest: &A{},
				src: &B{
					Field1: "hello",
					Field2: 42,
					Field3: 3.14,
				},
				omitExtra: true,
			},
			wantErr: false,
		},
		{
			name: "dest does not include all src fields",
			args: args{
				dest: &A{},
				src: &B{
					Field1: "hello",
					Field2: 42,
					Field3: 3.14,
				},
				omitExtra: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := assignStructFields(tt.args.dest, tt.args.src, tt.args.omitExtra); (err != nil) != tt.wantErr {
				t.Errorf("assignFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
