package monkey_test

import (
	"bytes"
	"github.com/bradleyjkemp/monkey"
	"github.com/stretchr/testify/require"
	"io"
	"sync"
	"testing"
)

func TestT(t *testing.T) {
	type matcher struct {
		matchFunc func(string,string) (bool, error)
		mu sync.Mutex
	}
	type testContext struct {
		match *matcher
	}
	type T struct {
		context *testContext
	}
	type common struct {
		w io.Writer
	}

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

	err := monkey.Patch(myT, shadow)
	require.NoError(t, err)

	parentShadow := &struct{
		common common
	}{
		common{
			w: &bytes.Buffer{},
		},
	}
	err = monkey.Patch(myT, parentShadow)
	require.NoError(t, err)

	myT.Run("muahahaha", func(t *testing.T) {
		t.Fatalf("this shouldn't fail my test")
	})

	require.True(t, myT.Failed())
}
