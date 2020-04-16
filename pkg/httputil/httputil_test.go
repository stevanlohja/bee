package httputil_test

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/ethersphere/bee/pkg/httputil"
)

func TestUnmarshal(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.vars != nil {
				t.Run("MuxVars", func(t *testing.T) {
					v, want := tc.setup()
					if err := httputil.UnmarshalMuxVars(tc.vars, v); err != nil {
						t.Fatal(err)
					}

					if !reflect.DeepEqual(v, want) {
						t.Errorf("got %+v, want %+v", v, want)
					}
				})
			}
			if tc.vars != nil {
				t.Run("URLValues", func(t *testing.T) {
					v, want := tc.setup()
					if err := httputil.UnmarshalURLValues(tc.vals, v); err != nil {
						t.Fatal(err)
					}

					if !reflect.DeepEqual(v, want) {
						t.Errorf("got %+v, want %+v", v, want)
					}
				})
			}
		})
	}
}

type testCase struct {
	name  string
	vars  map[string]string
	vals  url.Values
	setup func() (v, want interface{})
}

var testCases = []testCase{
	{
		name: "empty",
		vars: make(map[string]string),
		vals: url.Values{},
		setup: func() (v, want interface{}) {
			return struct{}{}, struct{}{}
		},
	},

	{
		name: "string",
		vars: map[string]string{"Str": "mux"},
		vals: url.Values{"Str": []string{"mux"}},
		setup: func() (v, want interface{}) {
			type t struct{ Str string }
			return &t{}, &t{Str: "mux"}
		},
	},
	{
		name: "string ptr",
		vars: map[string]string{"Str": "mux"},
		vals: url.Values{"Str": []string{"mux"}},
		setup: func() (v, want interface{}) {
			type t struct{ Str *string }
			return &t{}, &t{Str: stringPtr("mux")}
		},
	},

	{
		name: "int",
		vars: map[string]string{"Int": "2"},
		vals: url.Values{"Int": []string{"2"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int int }
			return &t{}, &t{Int: 2}
		},
	},
	{
		name: "int ptr",
		vars: map[string]string{"Int": "3"},
		vals: url.Values{"Int": []string{"3"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int *int }
			var i int = 3
			return &t{}, &t{Int: &i}
		},
	},
	{
		name: "int8",
		vars: map[string]string{"Int": "24"},
		vals: url.Values{"Int": []string{"24"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int int8 }
			return &t{}, &t{Int: 24}
		},
	},
	{
		name: "int8 ptr",
		vars: map[string]string{"Int": "34"},
		vals: url.Values{"Int": []string{"34"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int *int8 }
			var i int8 = 34
			return &t{}, &t{Int: &i}
		},
	},
	{
		name: "int16",
		vars: map[string]string{"Int": "24"},
		vals: url.Values{"Int": []string{"24"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int int16 }
			return &t{}, &t{Int: 24}
		},
	},
	{
		name: "int16 ptr",
		vars: map[string]string{"Int": "34"},
		vals: url.Values{"Int": []string{"34"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int *int16 }
			var i int16 = 34
			return &t{}, &t{Int: &i}
		},
	},
	{
		name: "int32",
		vars: map[string]string{"Int": "24"},
		vals: url.Values{"Int": []string{"24"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int int32 }
			return &t{}, &t{Int: 24}
		},
	},
	{
		name: "int32 ptr",
		vars: map[string]string{"Int": "34"},
		vals: url.Values{"Int": []string{"34"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int *int32 }
			var i int32 = 34
			return &t{}, &t{Int: &i}
		},
	},
	{
		name: "int64",
		vars: map[string]string{"Int": "24"},
		vals: url.Values{"Int": []string{"24"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int int64 }
			return &t{}, &t{Int: 24}
		},
	},
	{
		name: "int64 ptr",
		vars: map[string]string{"Int": "34"},
		vals: url.Values{"Int": []string{"34"}},
		setup: func() (v, want interface{}) {
			type t struct{ Int *int64 }
			var i int64 = 34
			return &t{}, &t{Int: &i}
		},
	},

	{
		name: "uint",
		vars: map[string]string{"UInt": "2"},
		vals: url.Values{"UInt": []string{"2"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt uint }
			return &t{}, &t{UInt: 2}
		},
	},
	{
		name: "uint ptr",
		vars: map[string]string{"UInt": "3"},
		vals: url.Values{"UInt": []string{"3"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt *uint }
			var i uint = 3
			return &t{}, &t{UInt: &i}
		},
	},
	{
		name: "uint8",
		vars: map[string]string{"UInt": "24"},
		vals: url.Values{"UInt": []string{"24"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt uint8 }
			return &t{}, &t{UInt: 24}
		},
	},
	{
		name: "uint8 ptr",
		vars: map[string]string{"UInt": "34"},
		vals: url.Values{"UInt": []string{"34"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt *uint8 }
			var i uint8 = 34
			return &t{}, &t{UInt: &i}
		},
	},
	{
		name: "uint16",
		vars: map[string]string{"UInt": "24"},
		vals: url.Values{"UInt": []string{"24"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt uint16 }
			return &t{}, &t{UInt: 24}
		},
	},
	{
		name: "uint16 ptr",
		vars: map[string]string{"UInt": "34"},
		vals: url.Values{"UInt": []string{"34"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt *uint16 }
			var i uint16 = 34
			return &t{}, &t{UInt: &i}
		},
	},
	{
		name: "uint32",
		vars: map[string]string{"UInt": "24"},
		vals: url.Values{"UInt": []string{"24"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt uint32 }
			return &t{}, &t{UInt: 24}
		},
	},
	{
		name: "uint32 ptr",
		vars: map[string]string{"UInt": "34"},
		vals: url.Values{"UInt": []string{"34"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt *uint32 }
			var i uint32 = 34
			return &t{}, &t{UInt: &i}
		},
	},
	{
		name: "uint64",
		vars: map[string]string{"UInt": "24"},
		vals: url.Values{"UInt": []string{"24"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt uint64 }
			return &t{}, &t{UInt: 24}
		},
	},
	{
		name: "uint64 ptr",
		vars: map[string]string{"UInt": "34"},
		vals: url.Values{"UInt": []string{"34"}},
		setup: func() (v, want interface{}) {
			type t struct{ UInt *uint64 }
			var i uint64 = 34
			return &t{}, &t{UInt: &i}
		},
	},

	{
		name: "float32",
		vars: map[string]string{"Float": "24.01"},
		vals: url.Values{"Float": []string{"24.01"}},
		setup: func() (v, want interface{}) {
			type t struct{ Float float32 }
			return &t{}, &t{Float: 24.01}
		},
	},
	{
		name: "float32 ptr",
		vars: map[string]string{"Float": "-34.84"},
		vals: url.Values{"Float": []string{"-34.84"}},
		setup: func() (v, want interface{}) {
			type t struct{ Float *float32 }
			var i float32 = -34.84
			return &t{}, &t{Float: &i}
		},
	},

	{
		name: "float64",
		vars: map[string]string{"Float": "-24.01"},
		vals: url.Values{"Float": []string{"-24.01"}},
		setup: func() (v, want interface{}) {
			type t struct{ Float float64 }
			return &t{}, &t{Float: -24.01}
		},
	},
	{
		name: "float64 ptr",
		vars: map[string]string{"Float": "34.84"},
		vals: url.Values{"Float": []string{"34.84"}},
		setup: func() (v, want interface{}) {
			type t struct{ Float *float64 }
			var i float64 = 34.84
			return &t{}, &t{Float: &i}
		},
	},

	{
		name: "slice string",
		vals: url.Values{"StringSlice": []string{"mux1", "mux2"}},
		setup: func() (v, want interface{}) {
			type t struct{ StringSlice []string }
			return &t{}, &t{StringSlice: []string{"mux1", "mux2"}}
		},
	},
	{
		name: "slice string ptr",
		vals: url.Values{"StringSlice": []string{"mux1", "mux2"}},
		setup: func() (v, want interface{}) {
			type t struct{ StringSlice []*string }
			return &t{}, &t{StringSlice: []*string{stringPtr("mux1"), stringPtr("mux2")}}
		},
	},

	{
		name: "slice int",
		vals: url.Values{"StringSlice": []string{"100", "120", "240"}},
		setup: func() (v, want interface{}) {
			type t struct{ StringSlice []int }
			return &t{}, &t{StringSlice: []int{100, 120, 240}}
		},
	},
	{
		name: "slice int ptr",
		vals: url.Values{"StringSlice": []string{"100", "120", "240"}},
		setup: func() (v, want interface{}) {
			type t struct{ StringSlice []*int }
			return &t{}, &t{StringSlice: []*int{intPtr(100), intPtr(120), intPtr(240)}}
		},
	},

	{
		name: "json tag name",
		vars: map[string]string{"my-name": "mux"},
		vals: url.Values{"my-name": []string{"mux"}},
		setup: func() (v, want interface{}) {
			type t struct {
				Str string `json:"my-name"`
			}
			return &t{}, &t{Str: "mux"}
		},
	},
	{
		name: "json tag name - blank",
		vars: map[string]string{"my-name": "mux"},
		vals: url.Values{"my-name": []string{"mux"}},
		setup: func() (v, want interface{}) {
			type t struct {
				Str string `json:""`
			}
			return &t{}, &t{Str: ""}
		},
	},
	{
		name: "json tag name - ignore",
		vars: map[string]string{"my-name": "mux"},
		vals: url.Values{"my-name": []string{"mux"}},
		setup: func() (v, want interface{}) {
			type t struct {
				Str string `json:""`
			}
			return &t{}, &t{Str: ""}
		},
	},
	{
		name: "json tag name - with omitempty",
		vars: map[string]string{"my-name": "mux"},
		vals: url.Values{"my-name": []string{"mux"}},
		setup: func() (v, want interface{}) {
			type t struct {
				Str string `json:"my-name,omitempty"`
			}
			return &t{}, &t{Str: "mux"}
		},
	},
}

func init() {
	for _, s := range []string{"", "1", "true", "on", "TRUE", "ON"} {
		testCases = append(testCases, newBoolTestCase(s, true))
		testCases = append(testCases, newBoolPtrTestCase(s, true))
	}
	for _, s := range []string{"0", "false", "off", "FALSE", "OFF"} {
		testCases = append(testCases, newBoolTestCase(s, false))
		testCases = append(testCases, newBoolPtrTestCase(s, false))
	}
}

func newBoolTestCase(s string, val bool) testCase {
	return testCase{
		name: "bool " + s,
		vars: map[string]string{"Bool": s},
		vals: url.Values{"Bool": []string{s}},
		setup: func() (v, want interface{}) {
			type t struct{ Bool bool }
			return &t{}, &t{Bool: val}
		},
	}
}

func newBoolPtrTestCase(s string, val bool) testCase {
	return testCase{
		name: "bool ptr " + s,
		vars: map[string]string{"Bool": s},
		vals: url.Values{"Bool": []string{s}},
		setup: func() (v, want interface{}) {
			type t struct{ Bool *bool }
			var b bool = val
			return &t{}, &t{Bool: &b}
		},
	}
}

func stringPtr(v string) *string { return &v }
func intPtr(v int) *int          { return &v }
