package monkey_test

import (
	"github.com/bradleyjkemp/monkey"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	actual := struct{
		isAdmin bool
	}{
		isAdmin: false,
	}

	shadow := struct{
		isAdmin bool
	}{
		isAdmin: true,
	}

	err := monkey.Patch(&actual, &shadow)
	require.NoError(t, err)
	require.Equal(t, true, actual.isAdmin)
}

func TestShallowRecursivePointers(t *testing.T) {
	actual := struct{
		isAdmin *bool
	}{}
	isAdmin := false
	actual.isAdmin = &isAdmin

	shadow := struct{
		isAdmin *bool
	}{}
	isAdminShadow := true
	shadow.isAdmin = &isAdminShadow

	err := monkey.Patch(&actual, &shadow)
	require.NoError(t, err)
	require.True(t, *actual.isAdmin)

	// pointer should have been overwritten so can change actual
	// by toggling shadow
	*shadow.isAdmin = false
	require.False(t, *actual.isAdmin)
}

func TestDeepRecursivePointers(t *testing.T) {
	type actualBool struct{
		bool bool
	}
	actual := struct{
		isAdmin *actualBool
	}{
		&actualBool{
			false,
		},
	}

	type shadowBool struct{
		bool bool
	}
	shadow := struct{
		isAdmin *shadowBool
	}{
		&shadowBool{
			true,
		},
	}

	err := monkey.Patch(&actual, &shadow)
	require.NoError(t, err)
	require.True(t, actual.isAdmin.bool)

	// pointer types didn't match so toggling shadow now shouldn't change actual
	shadow.isAdmin.bool = false
	require.True(t, actual.isAdmin.bool)
}


type flag struct {
	isSet bool
}
func (f *flag) set() {
	f.isSet = true
}
func (f *flag) get() bool {
	return f.isSet
}
type settable interface{
	set()
	get() bool
}
type shadowFlag struct {
	isSet bool
}
func (f *shadowFlag) set() {
	f.isSet = true
}
func (f *shadowFlag) get() bool {
	return f.isSet
}

func TestShallowInterface(t *testing.T) {
	actualFlag := &flag{}
	actual := struct{
		isAdmin settable
	}{
		actualFlag,
	}

	myFlag := &shadowFlag{}
	shadow := struct {
		// same type as actual so will just overwrite directly
		isAdmin settable
	}{
		myFlag,
	}

	err := monkey.Patch(&actual, &shadow)
	require.NoError(t, err)

	actual.isAdmin.set()
	require.True(t, myFlag.isSet)
	require.False(t, actualFlag.isSet)
	_, ok := actual.isAdmin.(*shadowFlag)
	require.True(t, ok)
}

func TestDeepInterface (t *testing.T) {
	actualFlag := &flag{}
	actual := struct{
		isAdmin settable
	}{
		actualFlag,
	}

	myFlag := &shadowFlag{true}
	shadow := struct {
		// different type so should recurse and edit the isSet field within actualFlag
		isAdmin *shadowFlag
	}{
		myFlag,
	}

	err := monkey.Patch(&actual, &shadow)
	require.NoError(t, err)

	require.True(t, actual.isAdmin.get())
	// should not have overwritten with a shadow flag, should have modified the real flag
	require.IsType(t, &flag{}, actual.isAdmin)
}
