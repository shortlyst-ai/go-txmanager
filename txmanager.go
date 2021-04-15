package txmanager

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type TxFn func(ctx context.Context) error

type TxManager interface {
	WithTransaction(ctx context.Context, txfn TxFn) error
}

type GormTxManager struct {
	db	*gorm.DB
}

// StartTxManager create TxManager with db
func StartTxManager(db *gorm.DB) TxManager {
	return &GormTxManager{db: db}
}

// WithTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `TxFn`
func (g *GormTxManager) WithTransaction(parentCtx context.Context, txfn TxFn) (err error) {
	tx := g.db.Begin()
	txCtx := setTxConn(parentCtx, tx)

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			tx.Rollback()
			logrus.Error(p)
			err = errors.New("panic happened because: " + fmt.Sprintf("%v", p))
		} else if err != nil {
			// error occurred, rollback
			tx.Rollback()
		} else {
			// all good, commit
			err = tx.Commit().Error
		}
	}()

	err = txfn(txCtx)
	return err
}