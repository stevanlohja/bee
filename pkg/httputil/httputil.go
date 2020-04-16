package httputil

import (
	"encoding"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func UnmarshalMuxVars(vars map[string]string, v interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	e := val.Type()
	if kind := e.Kind(); kind != reflect.Struct {
		return fmt.Errorf("struct type required, not %s", kind)
	}
	for i := 0; i < e.NumField(); i++ {
		name := fieldName(e.Field(i))
		if name == "" {
			continue
		}

		s, ok := vars[name]
		if !ok {
			continue
		}

		if err := setValue(val.Field(i), s); err != nil {
			return NewFieldError(name, err)
		}
	}
	return nil
}

func UnmarshalURLValues(q url.Values, v interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	e := val.Type()
	if kind := e.Kind(); kind != reflect.Struct {
		return fmt.Errorf("struct type required, not %s", kind)
	}
	for i := 0; i < e.NumField(); i++ {
		name := fieldName(e.Field(i))
		if name == "" {
			continue
		}

		values, ok := q[name]
		if !ok {
			continue
		}

		if err := setValue(val.Field(i), values...); err != nil {
			return NewFieldError(name, err)
		}
	}
	return nil
}

func fieldName(f reflect.StructField) string {
	name, set := f.Tag.Lookup("json")
	name = strings.TrimSuffix(name, ",omitempty")
	if name == "" || name == "-" {
		if set {
			return ""
		}
		return f.Name
	}
	return name
}

func setValue(v reflect.Value, s ...string) error {
	t := v.Type()

	kind := v.Kind()
	if kind != reflect.Ptr && t.Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText([]byte(s[0])); err != nil {
			return err
		}
		return nil
	}
	if v.IsNil() && kind != reflect.Slice {
		v.Set(reflect.New(t.Elem()))
	}
	if kind == reflect.Ptr {
		kind = v.Elem().Kind()
	}

	v = reflect.Indirect(v)

	switch kind {
	case reflect.Bool:
		switch strings.ToLower(s[0]) {
		case "", "1", "true", "on":
			v.SetBool(true)
		case "0", "false", "off":
			v.SetBool(false)
		default:
			return fmt.Errorf("invalid value %s", s[0])
		}

	case reflect.String:
		v.SetString(s[0])

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			return err
		}
		if v.OverflowInt(n) {
			return ErrOverflow
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s[0], 10, 64)
		if err != nil {
			return err
		}
		if v.OverflowUint(n) {
			return ErrOverflow
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s[0], v.Type().Bits())
		if err != nil {
			return err
		}
		if v.OverflowFloat(n) {
			return ErrOverflow
		}
		v.SetFloat(n)

	case reflect.Slice:
		l := len(s)
		sl := reflect.MakeSlice(t, l, l)
		for i := 0; i < l; i++ {
			if err := setValue(sl.Index(i), s[i]); err != nil {
				return fmt.Errorf("invalid value [%v] %s", i, s[i])
			}
		}
		v.Set(sl)

	default:
		return fmt.Errorf("unsupported field type %s", kind)
	}

	return nil
}
