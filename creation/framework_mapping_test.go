package creation

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFactoryMapping(t *testing.T) {
	instance := GetProviderInstance()
	require.NotNil(t, instance)
}
