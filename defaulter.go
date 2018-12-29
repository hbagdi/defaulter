package defaulter

import (
	"errors"
	"reflect"
)

var (
	errArgpNotPointer = errors.New("argp is not a pointer to a struct")
	errNilArgs        = errors.New("argp or def cannot be nil")
	errNotSameKind    = errors.New("argp and def are not of same kind")
)

// Set sets zero fields in the struct pointed by argp to
// the corresponding fields in def.
func Set(argp, def interface{}) error {
	d := reflect.ValueOf(def)
	d = reflect.Indirect(d)

	if argp == nil || def == nil {
		return errNilArgs
	}
	if reflect.TypeOf(argp).Kind() != reflect.Ptr {
		return errArgpNotPointer
	}

	if reflect.TypeOf(argp).Elem().String() != d.Type().String() {
		return errNotSameKind
	}

	v := reflect.ValueOf(argp).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		defaultVal := d.Field(i)
		setStructField(v.Field(i), defaultVal)
	}
	return nil
}

func setStructField(field reflect.Value, defaultVal reflect.Value) {
	if !field.CanSet() {
		return
	}
	if !defaultVal.IsValid() {
		return
	}

	if isEmptyValue(field) {
		switch field.Kind() {
		case reflect.Slice:
			ref := reflect.New(field.Type())
			ref.Elem().Set(reflect.MakeSlice(field.Type(), defaultVal.Len(), defaultVal.Len()))
			for i := 0; i < defaultVal.Len(); i++ {
				d := defaultVal.Index(i)
				s := ref.Elem().Index(i)
				setStructField(s, d)
			}
			field.Set(ref.Elem())
		case reflect.Map:
			ref := reflect.New(field.Type())
			ref.Elem().Set(reflect.MakeMap(field.Type()))
			keys := defaultVal.MapKeys()
			for _, key := range keys {
				refKey := reflect.New(key.Type())
				setStructField(refKey.Elem(), key)
				valueValue := defaultVal.MapIndex(key)
				refValue := reflect.New(valueValue.Type())
				setStructField(refValue.Elem(), valueValue)
				ref.Elem().SetMapIndex(refKey.Elem(), refValue.Elem())
			}
			field.Set(ref.Elem())
		case reflect.Struct:
		case reflect.Ptr:
			if !defaultVal.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
		default:
			field.Set(defaultVal)
		}
	}

	switch field.Kind() {
	case reflect.Ptr:
		setStructField(field.Elem(), defaultVal.Elem())
	case reflect.Struct:
		ref := reflect.New(field.Type())
		Set(ref.Interface(), defaultVal.Interface())
		field.Set(ref.Elem())
	}

	return
}

// isEmptyValue returns true if field is a Zero Value.
func isEmptyValue(field reflect.Value) bool {
	return reflect.DeepEqual(reflect.Zero(field.Type()).Interface(), field.Interface())
}
