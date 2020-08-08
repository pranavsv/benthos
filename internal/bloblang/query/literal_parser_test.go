package query

import (
	"testing"

	"github.com/Jeffail/benthos/v3/lib/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiteralParserErrors(t *testing.T) {
	tests := map[string]struct {
		input string
		err   string
	}{
		"bad object key": {
			input: `{5:"foo"}`,
			err:   `line 1 char 1: object keys must be strings, received: int64`,
		},
		"bad array element": {
			input: `[5,null,"unterminated string]`,
			err:   `line 1 char 30: required: expected end quote`,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := tryParse(test.input, false)
			assert.Equal(t, test.err, err.ErrorAtPosition([]rune(test.input)))
		})
	}
}

func TestLiteralParser(t *testing.T) {
	type easyMsg struct {
		content string
		meta    map[string]string
	}

	tests := map[string]struct {
		mapping string
		result  interface{}
		err     string
		value   *interface{}
	}{
		"basic map": {
			mapping: `{"foo":"bar"}`,
			result: map[string]interface{}{
				"foo": "bar",
			},
		},
		"dynamic map": {
			mapping: `{"foo":(5 + 5)}`,
			result: map[string]interface{}{
				"foo": float64(10),
			},
		},
		"dynamic map dynamic key": {
			mapping: `{("foobar".uppercase()):5}`,
			result: map[string]interface{}{
				"FOOBAR": int64(5),
			},
		},
		"dynamic map nested": {
			mapping: `{"foo":{"bar":(5 + 5)}}`,
			result: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": float64(10),
				},
			},
		},
		"dynamic array": {
			mapping: `["foo",(5 + 5),null]`,
			result: []interface{}{
				"foo", float64(10), nil,
			},
		},
		"dynamic array nested": {
			mapping: `["foo",[(5 + 5),"bar"],null]`,
			result: []interface{}{
				"foo", []interface{}{float64(10), "bar"}, nil,
			},
		},
		"bad array element": {
			mapping: `["foo",(5 + "not a number"),"bar"]`,
			err:     "expected number value, found string: not a number",
		},
		"bad object value": {
			mapping: `{"foo":(5 + "not a number")}`,
			err:     "failed to resolve 'foo' value: expected number value, found string: not a number",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := Parse([]rune(test.mapping))
			require.Nil(t, res.Err)
			require.Implements(t, (*Function)(nil), res.Payload)
			q := res.Payload.(Function)

			result, err := q.Exec(FunctionContext{
				Index: 0, MsgBatch: message.New(nil),
				Value: test.value,
			})
			if len(test.err) > 0 {
				assert.EqualError(t, err, test.err)
			} else {
				assert.Equal(t, test.result, result)
			}
		})
	}
}
