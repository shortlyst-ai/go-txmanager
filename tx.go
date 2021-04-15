package txmanager

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

const TxConnKey = "txConn"

func GetTxConn(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return nil
	}

	ctxVal := ctx.Value(TxConnKey)
	if ctxVal == nil {
		return nil
	}

	dbConn, ok := ctxVal.(*gorm.DB)
	if !ok {
		logrus.Warnf("Invalid type, want: *gorm.DB got: %T", ctxVal)
		return nil
	}
	return dbConn
}

func setTxConn(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, TxConnKey, db)
}