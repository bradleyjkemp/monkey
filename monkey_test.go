package monkey

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	actual := struct {
		isAdmin bool
	}{
		isAdmin: false,
	}

	shadow := struct{
		isAdmin bool
	}{
		isAdmin: true,
	}

	err := Patch(&actual, &shadow)
	require.NoError(t, err)
	spew.Dump(actual, shadow)
	require.Equal(t, actual.isAdmin, true)
}


func trueMatcher(_, _ string) (bool, error) {
	return true, nil
}

type T struct {
	context *testContext
}
type testContext struct {
	match *matcher
}
type matcher struct {
	matchFunc func(string,string) (bool, error)
	mu sync.Mutex
}
type common struct {
	w io.Writer
}

func TestT(t *testing.T) {
	myT := &testing.T{}

	shadow := &T{
		context: &testContext{
			match: &matcher{
				matchFunc: func(_,_ string) (bool, error) {
					return true, nil
				},
				mu: sync.Mutex{},
			},
		},
	}

	err := Patch(myT, shadow)
	require.NoError(t, err)

	fmt.Println("now patching parent field")

	parentShadow := &struct{
		common common
	}{
		common{
			w: &bytes.Buffer{},
		},
	}
	err = Patch(myT, parentShadow)
	require.NoError(t, err)

	spew.Dump(myT)

	myT.Run("heheh", func(t *testing.T) {
		t.Fatalf("oh noes")
	})

	require.True(t, myT.Failed())
}