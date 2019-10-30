package echis

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_bind(t *testing.T) {
	type test struct {
		binder Binder
		src    interface{}
		check  func(*testing.T, Binder)
		err    error
	}

	cases := map[string]test{
		"non struct type should error": {
			binder: viper.New(),
			src:    "foo",
			err:    ErrInvalidType,
		},
		"nil nested struct": {
			binder: viper.New(),
			src: struct {
				Foo *struct {
					Bar string `config:"bar"`
				} `config:"foo"`
			}{
				Foo: nil,
			},
			check: func(t *testing.T, b Binder) {
				t.Helper()
			},
		},
		"binds from struct": {
			binder: viper.New(),
			src: struct {
				Foo string `config:"foo"`
			}{
				Foo: "bar",
			},
			check: func(t *testing.T, b Binder) {
				t.Helper()

				os.Setenv("FOO", "bar")
				defer os.Clearenv()

				v, ok := b.(*viper.Viper)
				require.True(t, ok)

				assert.Equal(t, "bar", v.GetString("foo"))
			},
		},
		"binds from pointer to struct": {
			binder: viper.New(),
			src: &struct {
				Foo string `config:"foo"`
			}{
				Foo: "bar",
			},
			check: func(t *testing.T, b Binder) {
				t.Helper()

				os.Setenv("FOO", "bar")
				defer os.Clearenv()

				v, ok := b.(*viper.Viper)
				require.True(t, ok)

				assert.Equal(t, "bar", v.GetString("foo"))
			},
		},
		"nested struct": {
			binder: viper.New(),
			src: struct {
				Foo string `config:"foo"`
				Bar struct {
					Baz string `config:"baz"`
				} `config:"bar"`
			}{
				Foo: "bar",
				Bar: struct {
					Baz string `config:"baz"`
				}{
					Baz: "baz",
				},
			},
			check: func(t *testing.T, b Binder) {
				t.Helper()

				os.Setenv("BAR_BAZ", "baz")
				defer os.Clearenv()

				v, ok := b.(*viper.Viper)
				require.True(t, ok)

				assert.Equal(t, "baz", v.GetString("bar_baz"))
			},
		},
		"nested pointer struct": {
			binder: viper.New(),
			src: struct {
				Foo string `config:"foo"`
				Bar *struct {
					Baz string `config:"baz"`
				} `config:"bar"`
			}{
				Foo: "bar",
				Bar: &struct {
					Baz string `config:"baz"`
				}{
					Baz: "baz",
				},
			},
			check: func(t *testing.T, b Binder) {
				t.Helper()

				os.Setenv("BAR_BAZ", "baz")
				defer os.Clearenv()

				v, ok := b.(*viper.Viper)
				require.True(t, ok)

				assert.Equal(t, "baz", v.GetString("bar_baz"))
			},
		},
		"pointer pointer": {
			binder: viper.New(),
			src: &struct {
				Foo string `config:"foo"`
				Bar *struct {
					Baz string `config:"baz"`
				} `config:"bar"`
			}{
				Foo: "bar",
				Bar: &struct {
					Baz string `config:"baz"`
				}{
					Baz: "baz",
				},
			},
			check: func(t *testing.T, b Binder) {
				t.Helper()

				os.Setenv("FOO", "bar")
				os.Setenv("BAR_BAZ", "baz")
				defer os.Clearenv()

				v, ok := b.(*viper.Viper)
				require.True(t, ok)

				assert.Equal(t, "baz", v.GetString("bar_baz"))
				assert.Equal(t, "bar", v.GetString("foo"))
			},
		},
		"deep nested": {
			binder: viper.New(),
			src: &struct {
				Foo string `config:"foo"`
				Bar *struct {
					Baz  string `config:"baz"`
					Fizz struct {
						Buzz string `config:"buzz"`
					} `config:"fizz"`
				} `config:"bar"`
			}{
				Foo: "bar",
				Bar: &struct {
					Baz  string `config:"baz"`
					Fizz struct {
						Buzz string `config:"buzz"`
					} `config:"fizz"`
				}{
					Baz: "baz",
					Fizz: struct {
						Buzz string `config:"buzz"`
					}{
						Buzz: "fizz",
					},
				},
			},
			check: func(t *testing.T, b Binder) {
				t.Helper()

				os.Setenv("FOO", "bar")
				os.Setenv("BAR_BAZ", "baz")
				os.Setenv("BAR_FIZZ_BUZZ", "foo")
				defer os.Clearenv()

				v, ok := b.(*viper.Viper)
				require.True(t, ok)

				assert.Equal(t, "baz", v.GetString("bar_baz"))
				assert.Equal(t, "bar", v.GetString("foo"))
				assert.Equal(t, "foo", v.GetString("bar_fizz_buzz"))
			},
		},
	}

	for name, test := range cases {
		test := test
		t.Run(name, func(t *testing.T) {
			err := bind(test.binder, test.src, Options{
				TagName:   "config",
				Seperator: "_",
			})

			assert.Equal(t, test.err, err)

			if test.check != nil {
				test.check(t, test.binder)
			}
		})
	}
}
