package monkey

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
)

// Patch lets you overwrite unexported fields of structs even when:
// * the fields are deeply nested
// * the fields are of an unexported type
//
// This is done by creating a shadow struct of the same layout.
// Patch then scans through the struct to the patched and for each field:
// * If a field of the same name exists in the shadow then it will be patched.
// * If the matching field is of the same type then it is overwritten directly.
// * If the matching field is of a different type then the two types are patched recursively.
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
			// unaddressable so can't change use to change values
			return errors.New("cannot use unaddressable shadow")
		}

		shadow = shadow.Elem()
	}

	return patch(actual, shadow)
}

func patch(actual reflect.Value, shadow reflect.Value) error {
	switch actual.Kind() {
	// Indirections
	case reflect.Interface:
		return patchInterface(actual, shadow)
	case reflect.Ptr:
		return patchPtr(actual, shadow)

	// Collections
	case reflect.Struct:
		return patchStruct(actual, shadow)
	case reflect.Slice, reflect.Array:
		return patchSlice(actual, shadow)
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

	// Anything else should just be overwritable
	default:
		fmt.Println("patching primitive")
		return patchPrimitive(actual, shadow)
	}
}

func patchInterface(actual reflect.Value, shadow reflect.Value) error {
	if shadow.Type() == actual.Type() {
		// valid to just assign directly
		return unsafeSet(actual, shadow)
	}

	// assume we're meant to be patching the underlying type
	// TODO: add check here that actual.Elem is the same Kind as shadow
	return patch(actual.Elem(), shadow)
}

func patchStruct(actual reflect.Value, shadow reflect.Value) error {
	// assume this is a struct
	for i := 0; i < actual.NumField(); i++ {
		actualStructField := actual.Type().Field(i)
		fieldName := actualStructField.Name

		shadowField := shadow.FieldByName(fieldName)

		// check if matching field found in shadow
		if shadowField.IsValid() {

			if actualStructField.Type == shadowField.Type() {
				// fields are same type so overwrite directly
				return unsafeSet(actual.FieldByName(fieldName), shadowField)
			}

			// fields not same type so need to patch recursively
			err := patch(actual.FieldByName(fieldName), shadowField)
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
	if shadow.IsNil() {
		// no more overwriting to do
		return nil
	}

	if actual.IsNil() && !shadow.IsNil() {
		// need to create a new value for actual to point at
		pointee := reflect.New(actual.Type().Elem())
		actual = reflect.NewAt(actual.Type(), unsafe.Pointer(actual.UnsafeAddr())).Elem()
		actual.Set(pointee)
	}

	return patch(actual.Elem(), shadow.Elem())
}

func patchSlice(actual reflect.Value, shadow reflect.Value) error {
	// TODO: should we allow slices of different length here?
	if actual.Len() != shadow.Len() {
		return errors.New("cannot patch slices of different length")
	}

	for i := 0; i < actual.Len(); i++ {
		err := patch(actual.Index(i), shadow.Index(i))
		if err != nil {
			return errors.Wrapf(err, "index %v:", i)
		}
	}

	return nil
}

func unsafeSet(actual, shadow reflect.Value) error {
	actual = reflect.NewAt(actual.Type(), unsafe.Pointer(actual.UnsafeAddr())).Elem()
	shadow = reflect.NewAt(shadow.Type(), unsafe.Pointer(shadow.UnsafeAddr())).Elem()
	actual.Set(shadow)

	return nil
}
