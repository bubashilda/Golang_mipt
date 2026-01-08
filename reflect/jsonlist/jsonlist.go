//go:build !solution

package jsonlist

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

func Marshal(w io.Writer, slice interface{}) error {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return &json.UnsupportedTypeError{Type: reflect.TypeOf(slice)}
	}

	encoder := json.NewEncoder(w)
	for i := 0; i < v.Len(); i++ {
		if err := encoder.Encode(v.Index(i).Interface()); err != nil {
			return fmt.Errorf("error encoding element %d: %w", i, err)
		}
	}
	return nil
}

func Unmarshal(r io.Reader, slice interface{}) error {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return &json.UnsupportedTypeError{Type: reflect.TypeOf(slice)}
	}

	sliceValue := v.Elem()
	elemType := sliceValue.Type().Elem()

	decoder := json.NewDecoder(r)

	for decoder.More() {
		elemPtr := reflect.New(elemType)
		if err := decoder.Decode(elemPtr.Interface()); err != nil {
			return err
		}
		sliceValue = reflect.Append(sliceValue, elemPtr.Elem())
	}

	v.Elem().Set(sliceValue)
	return nil
}
