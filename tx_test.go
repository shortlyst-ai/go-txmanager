package txmanager_test

import (
	"context"
	"github.com/shortlyst-ai/go-txmanager"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetTxConn(t *testing.T) {
	t.Run("NilContext", func(t *testing.T) {
		db := txmanager.GetTxConn(nil)
		require.Nil(t, db)
	})

	t.Run("NilContextValue", func(t *testing.T) {
		db := txmanager.GetTxConn(context.Background())
		require.Nil(t, db)
	})

	t.Run("InvalidType", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), txmanager.TxConnKey, 1)
		db := txmanager.GetTxConn(ctx)
		require.Nil(t, db)
	})
}