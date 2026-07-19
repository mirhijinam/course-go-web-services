package main

import (
	"fmt"
	"reflect"
)

func i2s(data any, out any) error {
	_, err := parse(data, out)
	return err
}

func parse(data any, out any) (reflect.Value, error) {
	switch reflect.ValueOf(data).Kind() {
	case reflect.Map:
		return parseStruct(data, out)
	case reflect.Slice:
		return parseSlice(data, out)
	case reflect.Float64:
		return parseInt(data, out)
	case reflect.Bool:
		return parseBool(data, out)
	case reflect.String:
		return parseString(data, out)
	default:
		return reflect.Value{}, nil
	}
}

func parseStruct(data any, out any) (reflect.Value, error) {
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Pointer {
		return reflect.Value{}, fmt.Errorf("error struct 1: %v", v.Kind())
	}

	o := v.Elem()
	if o.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("error struct 2: %v", o.Kind())
	}

	m := data.(map[string]any)
	for k, v := range m {
		f := o.FieldByName(k)
		fAddr := f.Addr()

		_, err := parse(v, fAddr.Interface())
		if err != nil {
			return reflect.Value{}, err
		}
	}

	return o, nil
}

func parseInt(data any, out any) (reflect.Value, error) {
	o := reflect.ValueOf(out).Elem()
	if o.Kind() != reflect.Int {
		return reflect.Value{}, fmt.Errorf("error int: %v", o.Kind())
	}
	o.SetInt(int64(data.(float64)))
	return o, nil
}

func parseBool(data any, out any) (reflect.Value, error) {
	o := reflect.ValueOf(out).Elem()
	if o.Kind() != reflect.Bool {
		return reflect.Value{}, fmt.Errorf("error bool: %v", o.Kind())
	}
	o.SetBool(data.(bool))
	return o, nil
}

func parseString(data any, out any) (reflect.Value, error) {
	o := reflect.ValueOf(out).Elem()
	if o.Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf("error string: %v", o.Kind())
	}
	o.SetString(data.(string))
	return o, nil
}

func parseSlice(data any, out any) (reflect.Value, error) {
	d := reflect.ValueOf(data)

	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Pointer {
		return reflect.Value{}, fmt.Errorf("error slice 1: %v", v.Kind())
	}

	o := v.Elem()
	if o.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf("error slice 2: %v", o.Kind())
	}

	for i := range d.Len() {
		dd := d.Index(i)
		oo := reflect.New(v.Type().Elem().Elem())

		x, err := parse(dd.Interface(), oo.Interface())
		if err != nil {
			return reflect.Value{}, err
		}

		o = reflect.Append(o, x)
	}

	v.Elem().Set(o)
	return v, nil
}
