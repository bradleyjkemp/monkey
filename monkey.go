package monkey

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
)

func Patch(actualI interface{}, shadowI interface{}) error {
	actual := reflect.ValueOf(actualI)
	shadow := reflect.ValueOf(shadowI)
	if !actual.CanAddr() || !shadow.CanAddr() {
		if actual.Kind() != reflect.Ptr && actual.Kind() != reflect.Interface {
			// unaddressable so can't change values
			return errors.New("cannot patch unaddressable value")
		}

		actual = actual.Elem()

		if shadow.Kind() != reflect.Ptr && shadow.Kind() != reflect.Interface {
			// unaddressable so can't change values
			return errors.New("cannot use unaddressable patch")
		}

		shadow = shadow.Elem()
	}

	return patch(actual, shadow)
}

func patch(actual reflect.Value, shadow reflect.Value) error {
	fmt.Println("patching", actual, shadow)
	switch actual.Kind() {
	case reflect.Interface:
		// no use recursing inside
		return unsafeSet(actual, shadow)
	case reflect.Ptr:
		return patchPtr(actual, shadow)
	// Collections
	case reflect.Struct:
		fmt.Println("patching struct")
		return patchStruct(actual, shadow)
	//case reflect.Slice, reflect.Array:
	//	return m.mapSlice(iVal, parentID, inlineable)
	//case reflect.Map:
	//	return m.mapMap(iVal, parentID, inlineable)
	//
	//// Simple types
	//case reflect.Bool:
	//	return patchPrimitive(actual, shadow)
	//case reflect.String:
	//	return m.mapString(iVal, inlineable)
	//case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
	//	return m.mapInt(iVal, inlineable)
	//case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
	//	return m.mapUint(iVal, inlineable)

	// If we've missed anything then just skip it
	default:
		fmt.Println("patching primitive")
		return patchPrimitive(actual, shadow)
	}
}

func patchStruct(actual reflect.Value, shadow reflect.Value) error {
	// assume this is a struct
	for i := 0; i<actual.NumField(); i++ {
		actualStructField := actual.Type().Field(i)
		fieldName := actualStructField.Name
		fmt.Println("examining field", fieldName)

		patchField := shadow.FieldByName(fieldName)
		if patchField.IsValid() {
			patchStructField, _ := shadow.Type().FieldByName(fieldName)
			fieldTags := patchStructField.Tag

			if val, ok := fieldTags.Lookup("monkey"); ok && val == "shallow"{
				fmt.Println("patching field", fieldName, "shallow")
				return unsafeSet(actual.FieldByName(fieldName), patchField)
			}

			fmt.Println("patching field", fieldName, "recursively")
			err := patch(actual.FieldByName(fieldName), patchField)
			if err != nil {
				return errors.Wrap(err, fieldName)
			}
		}
	}

	return nil
}

func patchPrimitive(actual reflect.Value, shadow reflect.Value) error {
	return unsafeSet(actual, shadow)
}

func patchPtr(actual reflect.Value, shadow reflect.Value) error {
	fmt.Println("patching ptr")
	if shadow.IsNil() {
		// no more overwriting to do
		return nil
	}

	if actual.IsNil() && !shadow.IsNil() {
		// need to create a new value for actual to point at
		fmt.Println(actual.Type().Elem())
		pointee := reflect.New(actual.Type().Elem())
		actual = reflect.NewAt(actual.Type(), unsafe.Pointer(actual.UnsafeAddr())).Elem()
		actual.Set(pointee)
	}

	return patch(actual.Elem(), shadow.Elem())
}

func unsafeSet(actual, shadow reflect.Value) error {
	fmt.Println("unsafe set", actual, shadow)
	actual = reflect.NewAt(actual.Type(), unsafe.Pointer(actual.UnsafeAddr())).Elem()
	shadow = reflect.NewAt(shadow.Type(), unsafe.Pointer(shadow.UnsafeAddr())).Elem()
	actual.Set(shadow)

	return nil
}