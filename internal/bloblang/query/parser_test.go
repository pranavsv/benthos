package query

import (
	"testing"

	"github.com/Jeffail/benthos/v3/internal/bloblang/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctionParserErrors(t *testing.T) {
	tests := map[string]struct {
		input      string
		err        string
		deprecated bool
	}{
		"bad function": {
			input:      `not a function`,
			deprecated: true,
			err:        `line 1 char 4: expected function arguments`,
		},
		"bad function 2": {
			input: `not_a_function()`,
			err:   `line 1 char 1: unrecognised function 'not_a_function'`,
		},
		"bad args 2": {
			input:      `json("foo`,
			deprecated: true,
			err:        `line 1 char 10: required: expected end quote`,
		},
		"bad args 3": {
			input: `json(`,
			err:   `line 1 char 6: required: expected boolean, number, quoted string, or query`,
		},
		"bad args 4": {
			input: `json(0,`,
			err:   `line 1 char 8: required: expected boolean, number, quoted string, or query`,
		},
		"bad args 5": {
			input:      `json`,
			deprecated: true,
			err:        `line 1 char 5: expected function arguments`,
		},
		"bad args 7": {
			input: `json(5)`,
			err:   `line 1 char 1: expected string argument, received int64`,
		},
		"bad args 8": {
			input: `json(false)`,
			err:   `line 1 char 1: expected string argument, received bool`,
		},
		"bad operators": {
			input: `json("foo") + `,
			err:   `line 1 char 15: expected query`,
		},
		"bad expression": {
			input: `(json("foo") `,
			err:   `line 1 char 14: required: expected closing bracket`,
		},
		"bad expression 2": {
			input: `(json("foo") + `,
			err:   `line 1 char 16: expected query`,
		},
		"bad expression 3": {
			input: `(json("foo") + meta("bar") `,
			err:   `line 1 char 28: required: expected closing bracket`,
		},
		"bad method": {
			input: `json("foo").not_a_thing()`,
			err:   `line 1 char 13: unrecognised method 'not_a_thing'`,
		},
		"bad method 2": {
			input:      `json("foo").not_a_thing()`,
			deprecated: true,
			err:        `line 1 char 13: unrecognised method 'not_a_thing'`,
		},
		"bad method args 2": {
			input: `json("foo").from(`,
			err:   `line 1 char 18: required: expected boolean, number, quoted string, or query`,
		},
		"bad method args 3": {
			input: `json("foo").from()`,
			err:   `line 1 char 13: expected 1 arguments, received: 0`,
		},
		"bad method args 4": {
			input: `json("foo").from("nah")`,
			err:   `line 1 char 13: expected int argument, received string`,
		},
		"bad map args": {
			input: `json("foo").map()`,
			err:   `line 1 char 13: expected 1 arguments, received: 0`,
		},
		"gibberish": {
			input: `json("foo").(=)`,
			err:   `line 1 char 14: required: expected query`,
		},
		"gibberish 2": {
			input: `json("foo").(1 + )`,
			err:   `line 1 char 18: required: expected query`,
		},
		"bad match": {
			input: `match json("foo")`,
			err:   `line 1 char 18: required: expected {`,
		},
		"bad match 2": {
			input: `match json("foo") what is this?`,
			err:   `line 1 char 19: required: expected {`,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := tryParse(test.input, test.deprecated)
			require.NotNil(t, err)
			assert.Equal(t, test.err, err.ErrorAtPosition([]rune(test.input)))
		})
	}
}

func TestFunctionParserLimits(t *testing.T) {
	tests := map[string]struct {
		input      string
		remaining  string
		deprecated bool
	}{
		"nothing": {
			input:     `json("foo") + meta("bar")`,
			remaining: ``,
		},
		"space before": {
			input:     `   json("foo") + meta("bar")`,
			remaining: ``,
		},
		"space before 2": {
			input:     `   json("foo")   +    meta("bar")`,
			remaining: ``,
		},
		"unfinished comment": {
			input:     `json("foo") + meta("bar") # Here's a comment`,
			remaining: ` # Here's a comment`,
		},
		"extra text": {
			input:     `json("foo") and this`,
			remaining: ` and this`,
		},
		"extra text 2": {
			input:     `json("foo") + meta("bar") and this`,
			remaining: ` and this`,
		},
		"extra text 3": {
			input:     `json("foo")+meta("bar")and this`,
			remaining: `and this`,
		},
		"extra text 4": {
			input:     `json("foo")+meta("bar")         and this`,
			remaining: `         and this`,
		},
		"squiggly bracket": {
			input:     `json("foo")}`,
			remaining: `}`,
		},
		"normal bracket": {
			input:     `json("foo"))`,
			remaining: `)`,
		},
		"normal bracket 2": {
			input:     `json("foo"))))`,
			remaining: `)))`,
		},
		"normal bracket 3": {
			input:     `json("foo")) + json("bar")`,
			remaining: `) + json("bar")`,
		},
		"path literals": {
			input:     `this.foo bar baz`,
			remaining: ` bar baz`,
		},
		"path literals 2": {
			input:     `this.foo . bar baz`,
			remaining: ` . bar baz`,
		},
		"brackets at root": {
			input:     `(json().foo | "fallback").from_all()`,
			remaining: ``,
		},
		"brackets after root": {
			input:     `this.root.(json().foo | "fallback").from_all()`,
			remaining: ``,
		},
		"brackets after root 2": {
			input:     `this.root.(json().foo | "fallback").from_all().bar.baz`,
			remaining: ``,
		},
		"this at root": {
			input:     `this.foo.bar and then this`,
			remaining: ` and then this`,
		},
		"path literal at root": {
			input:     `foo.bar and then this`,
			remaining: ` and then this`,
		},
		"match expression": {
			input: `match null {
	"foo" == "bar" => "baz"
	5 > 10 => "or this"
}
not this`,
			remaining: "\nnot this",
		},
		"operators and line breaks": {
			input: `(5 * 8) +
	6 -
	5 and also this`,
			remaining: " and also this",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res parser.Result
			if test.deprecated {
				res = ParseDeprecated([]rune(test.input))
			} else {
				res = Parse([]rune(test.input))
			}
			require.Nil(t, res.Err)
			assert.Equal(t, test.remaining, string(res.Remaining))
		})
	}
}
