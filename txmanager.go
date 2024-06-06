package txmanager

import (
	"context"
	"errors"
	"fmt"
	"runtime"

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

	//execute txfn in background and communicates errors through a channel
	errCh := make(chan error)
	go func() {
		var errFunc error
		defer func() {
			if p := recover(); p != nil {
				logrus.Errorf("stack trace: %s", g.printStackTrace())
				errFunc = fmt.Errorf("panic recovered: %v", p)
			}
			errCh <- errFunc
		}()
		errFunc = txfn(txCtx)
	}()

	//wait for context cancelled or fn is success(err=nil) or failed(err=!nil)
	select {
	case <-parentCtx.Done():
		return parentCtx.Err()
	case err := <-errCh:
		return err
	}
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

func (g *GormTxManager) printStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
