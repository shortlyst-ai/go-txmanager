package txmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	gormv2 "gorm.io/gorm"
)

type TxFn func(ctx context.Context) error

type TxManager interface {
	WithTransaction(ctx context.Context, txfn TxFn) error
}

type GormTxManager struct {
	db *gorm.DB
}

type GormV2TxManager struct {
	db *gormv2.DB
}

// StartTxManager create TxManager with db
func StartTxManager(db *gorm.DB) TxManager {
	return &GormTxManager{db: db}
}

// NewGormTxManager create TxManagerGormV2 with dbv2
func NewGormTxManager(db *gormv2.DB) TxManager {
	return &GormV2TxManager{db: db}
}

// WithTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `TxFn`
func (g *GormTxManager) WithTransaction(parentCtx context.Context, txfn TxFn) (err error) {
	tx := g.db.Begin()
	txCtx := setTxConn(parentCtx, tx)

	isCtxCancelled := false
	goroutineChannel := make(chan bool)
	defer close(goroutineChannel)
	go func() {
		select {
		case <-txCtx.Done():
			isCtxCancelled = true
		case <-goroutineChannel:
			// to kill goroutine if transaction finished
			break
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			tx.Rollback()
			logrus.Error(p)
			err = errors.New("panic happened because: " + fmt.Sprintf("%v", p))
		} else if err != nil || isCtxCancelled {
			// error occurred, rollback
			tx.Rollback()
			if err == nil {
				err = errors.New("context cancelled")
			}
			return
		} else {
			// all good, commit
			err = tx.Commit().Error
		}
		goroutineChannel <- true
	}()

	err = txfn(txCtx)
	return err
}

func (g *GormV2TxManager) WithTransaction(parentCtx context.Context, txfn TxFn) (err error) {
	tx := g.db.Begin()
	txCtx := setTxConnV2(parentCtx, tx)

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
