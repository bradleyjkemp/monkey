package monkey_test

import (
	"github.com/bradleyjkemp/monkey"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUnaddressableActual(t *testing.T) {
	a := struct{}{}
	err := monkey.Patch(a, nil)
	require.Error(t, err, "actual must be passed by pointer/addressable value")
}

func TestUnaddressableShadow(t *testing.T) {
	a := struct{}{}
	err := monkey.Patch(&a, nil)
	require.Error(t, err, "shadow must be passed by pointer/addressable value")
}

