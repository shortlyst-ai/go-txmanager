package txmanager

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	gormv2 "gorm.io/gorm"
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

func GetTxConnV2(ctx context.Context) *gormv2.DB {
	if ctx == nil {
		return nil
	}

	ctxVal := ctx.Value(TxConnKey)
	if ctxVal == nil {
		return nil
	}

	dbConn, ok := ctxVal.(*gormv2.DB)
	if !ok {
		logrus.Warnf("Invalid type, want: *gormv2.DB got: %T", ctxVal)
		return nil
	}
	return dbConn
}

func setTxConnV2(ctx context.Context, db *gormv2.DB) context.Context {
	return context.WithValue(ctx, TxConnKey, db)
}
