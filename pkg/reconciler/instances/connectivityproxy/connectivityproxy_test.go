package connectivityproxy

import (
	"testing"

	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	t.Run("Should register Connectivity Proxy reconciler", func(t *testing.T) {
		reconciler, err := service.GetReconciler("connectivity-proxy")
		require.NoError(t, err)
		require.NotNil(t, reconciler)
	})
}
