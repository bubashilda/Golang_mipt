//go:build !solution

package reversemap

import "reflect"

func ReverseMap(forward interface{}) interface{} {
	mp := reflect.ValueOf(forward)

	ansType := reflect.MapOf(mp.Type().Elem(), mp.Type().Key())
	ans := reflect.MakeMap(ansType)

	iter := mp.MapRange()
	for iter.Next() {
		ans.SetMapIndex(iter.Value(), iter.Key())
	}

	return ans.Interface()
}
