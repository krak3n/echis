package echis

import (
	"reflect"
	"strings"
)

const (
	ErrInvalidType Error = iota + 1
	ErrNilBinder
)

const (
	DefaultSeperator = "_"
	DefaultTagName   = "mapstructure"
)

// Error is a custom error returned by Bind to indicate the error
type Error uint8

func (e Error) Error() string {
	switch e {
	case ErrInvalidType:
		return "invalid type: must be struct or pointer to a struct"
	case ErrNilBinder:
		return "nil binder"
	default:
		return "unknown error"
	}
}

// A Binder binds environment variables
type Binder interface {
	BindEnv(...string) error
}

// Options configures how binding occurs
type Options struct {
	Seperator string
	TagName   string
}

// An Option function updates an Options instance
type Option func(*Options)

// WithTagName configues the seperator used to join Viper look up paths
func WithSeperator(s string) Option {
	return func(o *Options) {
		o.Seperator = s
	}
}

// WithTagName configures the struct tag name used to walk the data structure
func WithTagName(n string) Option {
	return func(o *Options) {
		o.TagName = n
	}
}

// Bind walks the given source to build viper lookup paths from struct tags, by
// default it assumes the same struct tag as you would use for viper.Unmarshal
// but you can set a different tag name if you need using the WithTagName option.
// If you have also set Viper to use a different seperator from _ you can also set
// that using the WithSeperator option.
func Bind(binder Binder, src interface{}, opts ...Option) error {
	o := &Options{
		Seperator: DefaultSeperator,
		TagName:   DefaultTagName,
	}

	for _, opt := range opts {
		opt(o)
	}

	return bind(binder, src, *o)
}

func bind(b Binder, src interface{}, opts Options, parts ...string) error {
	if b == nil {
		return ErrNilBinder
	}

	v := reflect.ValueOf(src)
	t := reflect.TypeOf(src)

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v = reflect.New(t.Elem())
		}
		return bind(b, v.Elem().Interface(), opts, parts...)
	case reflect.Struct:
	default:
		return ErrInvalidType
	}

	// Loop over the fields on the struct, look for the struct tag configured in the options and
	// extract the tag value, use this value to pass to vipers's BindEnv method
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)

		tv, ok := t.Field(i).Tag.Lookup(opts.TagName)
		if !ok {
			continue
		}

		p := append(parts, tv)

		// Recursively call bind if the value is a nested struct
		if fv.Kind() == reflect.Struct || fv.Kind() == reflect.Ptr {
			if err := bind(b, fv.Interface(), opts, p...); err != nil {
				return err
			}
		}

		// Bind the environment variable
		if err := b.BindEnv(strings.Join(p, opts.Seperator)); err != nil {
			return err
		}
	}

	return nil
}
