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

type Binder interface {
	BindEnv(...string) error
}

type Options struct {
	Seperator string
	TagName   string
}

type Option func(*Options)

func WithSeperator(s string) Option {
	return func(o *Options) {
		o.Seperator = s
	}
}

func WithTagName(n string) Option {
	return func(o *Options) {
		o.TagName = n
	}
}

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
