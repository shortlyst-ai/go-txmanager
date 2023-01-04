package txmanager_test

import (
	"context"
	"testing"

	"github.com/shortlyst-ai/go-txmanager"
	"github.com/stretchr/testify/require"
)

func TestGetTxConn(t *testing.T) {
	t.Run("NilContext", func(t *testing.T) {
		db := txmanager.GetTxConn(nil)
		require.Nil(t, db)

		dbv2 := txmanager.GetTxConnV2(nil)
		require.Nil(t, dbv2)
	})

	t.Run("NilContextValue", func(t *testing.T) {
		db := txmanager.GetTxConn(context.Background())
		require.Nil(t, db)

		dbv2 := txmanager.GetTxConnV2(context.Background())
		require.Nil(t, dbv2)
	})

	t.Run("InvalidType", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), txmanager.TxConnKey, 1)
		db := txmanager.GetTxConn(ctx)
		require.Nil(t, db)

		dbv2 := txmanager.GetTxConnV2(ctx)
		require.Nil(t, dbv2)
	})
}
