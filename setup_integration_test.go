package txmanager_test

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/shortlyst-ai/go-txmanager"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var dbName string

func getMysql() *gorm.DB {
	dbName = os.Getenv("DB_DATABASE")
	dbURL := getMysqlUrl(dbName)
	db, err := gorm.Open("mysql", dbURL)
	if err != nil {
		panic(fmt.Sprintf("error: %v for %v", err.Error(), dbURL))
	}

	db.Delete(&author{})
	db.Delete(&book{})
	db.Delete(&authorBook{})

	return db
}

func closeDB(t *testing.T, db *gorm.DB) {
	db.Delete(&author{})
	db.Delete(&book{})
	db.Delete(&authorBook{})

	err := db.Close()
	require.NoError(t, err)
}

func resetDB(db *gorm.DB) error {
	models := []interface{}{
		book{},
		author{},
		authorBook{},
	}

	for _, v := range models {
		tableName := db.NewScope(v).TableName()
		err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName)).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func getMysqlUrl(dbName string) string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, pass, host, port, dbName)
}

func stringPointer(str string) *string {
	return &str
}

type author struct {
	ID		*int 	`gorm:"primaryKey;autoIncrement" json:"id"`
	Name	*string	`gorm:"unique" json:"name"`
}

type book struct {
	ID		*int 	`gorm:"primaryKey;autoIncrement" json:"id"`
	Name	*string	`json:"name"`
}

type authorBook struct {
	ID			*int 	`json:"id"`
	AuthorID	*int	`json:"author_id"`
	BookID		*int	`json:"book_id"`
}

type authorRepository interface {
	AddAuthor(ctx context.Context, author author) (*author, error)
}

type bookRepository interface {
	AddBook(ctx context.Context, book book) (*book, error)
}

type authorBookRepository interface {
	LinkAuthorBook(ctx context.Context, author author, book book) (*authorBook, error)
}

type repository struct {
	db *gorm.DB
}

func newAuthorRepository(db *gorm.DB) authorRepository {
	return repository{
		db: db,
	}
}

func newBookRepository(db *gorm.DB) bookRepository {
	return repository{
		db: db,
	}
}

func newAuthorBookRepository(db *gorm.DB) authorBookRepository {
	return repository{
		db: db,
	}
}

func (r repository) AddAuthor(ctx context.Context, author author) (res *author, err error) {
	var tx *gorm.DB

	if txConn := txmanager.GetTxConn(ctx); txConn != nil {
		tx = txConn
	}

	if tx == nil {
		tx = r.db.Begin()
		defer func() {
			if err != nil {
				tx.Rollback()
				return
			}
			tx.Commit()
		}()
	}

	err = tx.Create(&author).Error

	if err != nil {
		return nil, err
	}

	res = &author

	return res, nil
}

func (r repository) AddBook(ctx context.Context, book book) (res *book, err error) {
	var tx *gorm.DB

	if txConn := txmanager.GetTxConn(ctx); txConn != nil {
		tx = txConn
	}

	if tx == nil {
		tx = r.db.Begin()
		defer func() {
			if err != nil {
				tx.Rollback()
				return
			}
			tx.Commit()
		}()
	}

	err = tx.Create(&book).Error

	if err != nil {
		return nil, err
	}

	res = &book

	return res, nil
}

func (r repository) LinkAuthorBook(ctx context.Context, author author, book book) (res *authorBook, err error) {
	var tx *gorm.DB

	if txConn := txmanager.GetTxConn(ctx); txConn != nil {
		tx = txConn
	}

	if tx == nil {
		tx = r.db.Begin()
		defer func() {
			if err != nil {
				tx.Rollback()
				return
			}
			tx.Commit()
		}()
	}

	aB := authorBook{
		AuthorID: author.ID,
		BookID:   book.ID,
	}

	err = tx.Create(&aB).Error

	if err != nil {
		return nil, err
	}

	res = &aB

	return res, nil
}