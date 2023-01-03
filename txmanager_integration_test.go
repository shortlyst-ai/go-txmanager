package txmanager_test

import (
	"context"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/shortlyst-ai/go-txmanager"
	"github.com/stretchr/testify/require"
	gormv2 "gorm.io/gorm"
)

func TestTxManager_WithTransaction(t *testing.T) {
	db := getMysql()
	db.AutoMigrate(&author{}, &book{}, &authorBook{})
	defer closeDB(t, db)

	dbv2 := getMysqlV2()
	dbv2.AutoMigrate(&author{}, &book{}, &authorBook{})
	defer closeDBV2(t, dbv2)

	author1 := author{
		Name: stringPointer("john"),
	}
	book1 := book{
		Name: stringPointer("Math"),
	}

	repoAuthor := newAuthorRepository(db, dbv2)
	repoBook := newBookRepository(db, dbv2)
	repo := newAuthorBookRepository(db, dbv2)

	t.Run("AcrossRepoTx_Success", func(t *testing.T) {
		// reset db after test
		defer func(db *gorm.DB) {
			err := resetDB(db)
			require.NoError(t, err)
		}(db)

		// start TxManager
		txManager := txmanager.StartTxManager(db)

		// doing transaction, add author & book, and then link the author and book
		transaction := func(ctx context.Context) error {
			authorSaved, err := repoAuthor.AddAuthor(ctx, author1)

			if err != nil {
				return err
			}

			bookSaved, err := repoBook.AddBook(ctx, book1)

			if err != nil {
				return err
			}

			_, err = repo.LinkAuthorBook(ctx, *authorSaved, *bookSaved)

			if err != nil {
				return err
			}

			return nil
		}

		err := txManager.WithTransaction(context.Background(), transaction)
		require.NoError(t, err)

		// validate data
		var authorResult author
		err = db.Model(&author{}).Find(&authorResult, "name = ?", "john").Error
		require.NoError(t, err)
		require.Equal(t, author1.Name, authorResult.Name)

		var bookResult book
		err = db.Model(&book{}).Find(&bookResult, "name = ?", "Math").Error
		require.NoError(t, err)
		require.Equal(t, book1.Name, bookResult.Name)

		var aBResult authorBook
		err = db.Model(&authorBook{}).Find(&aBResult, "author_id = ?", authorResult.ID).Error
		require.NoError(t, err)
		require.Equal(t, *authorResult.ID, *aBResult.AuthorID)
		require.Equal(t, *bookResult.ID, *aBResult.BookID)
	})

	t.Run("AcrossRepoTxFailed_ThenRollback", func(t *testing.T) {
		// reset db after test
		defer func(db *gorm.DB) {
			err := resetDB(db)
			require.NoError(t, err)
		}(db)

		// start TxManager
		txManager := txmanager.StartTxManager(db)
		book2 := book{
			Name: stringPointer("Biology"),
		}

		// save author1 data first, to trigger error on below tx
		err := db.Create(&author1).Error
		require.NoError(t, err)

		// doing transaction, add author & book, and then link the author and book
		transaction := func(ctx context.Context) error {
			// saved book2
			bookSaved, err := repoBook.AddBook(ctx, book2)

			if err != nil {
				return err
			}

			// saving author1 again will trigger error unique name, because we set on author name must be unique
			authorSaved, err := repoAuthor.AddAuthor(ctx, author1)

			if err != nil {
				return err
			}

			_, err = repo.LinkAuthorBook(ctx, *authorSaved, *bookSaved)

			if err != nil {
				return err
			}

			return nil
		}

		err = txManager.WithTransaction(context.Background(), transaction)
		require.Error(t, err)

		// validate data
		var bookResult book
		err = db.Model(&book{}).Find(&bookResult, "name = ?", "Biology").Error
		require.Error(t, err)
		require.Nil(t, bookResult.ID)
	})

	t.Run("Panic_ThenRollback", func(t *testing.T) {
		// reset db after test
		defer func(db *gorm.DB) {
			err := resetDB(db)
			require.NoError(t, err)
		}(db)

		// start TxManager
		txManager := txmanager.StartTxManager(db)
		book2 := book{
			Name: stringPointer("Biology"),
		}

		// doing transaction, add author & book, and then link the author and book
		transaction := func(ctx context.Context) error {
			// saved book2
			bookSaved, err := repoBook.AddBook(ctx, book2)

			if err != nil {
				return err
			}

			// test have nil pointer value to book
			// when get the value from test, it will trigger panic nil pointer
			var test *book
			failing := *test

			failing.Name = stringPointer("test_panic")

			authorSaved, err := repoAuthor.AddAuthor(ctx, author1)

			if err != nil {
				return err
			}

			_, err = repo.LinkAuthorBook(ctx, *authorSaved, *bookSaved)

			if err != nil {
				return err
			}

			return nil
		}

		err := txManager.WithTransaction(context.Background(), transaction)
		require.Error(t, err)

		// validate data
		var bookResult book
		err = db.Model(&book{}).Find(&bookResult, "name = ?", "Biology").Error
		require.Error(t, err)
		require.Nil(t, bookResult.ID)
	})

	t.Run("AcrossRepoV2Tx_Success", func(t *testing.T) {
		// reset db after test
		defer func(db *gormv2.DB) {
			err := resetDBV2(db)
			require.NoError(t, err)
		}(dbv2)

		// start TxManager
		txManager := txmanager.StartTxManagerGormV2(dbv2)

		// doing transaction, add author & book, and then link the author and book
		transaction := func(ctx context.Context) error {
			authorSaved, err := repoAuthor.AddAuthorV2(ctx, author1)
			if err != nil {
				return err
			}

			bookSaved, err := repoBook.AddBookV2(ctx, book1)
			if err != nil {
				return err
			}

			_, err = repo.LinkAuthorBookV2(ctx, *authorSaved, *bookSaved)
			if err != nil {
				return err
			}

			return nil
		}

		err := txManager.WithTransaction(context.Background(), transaction)
		require.NoError(t, err)

		// validate data
		var authorResult author
		err = db.Model(&author{}).Find(&authorResult, "name = ?", "john").Error
		require.NoError(t, err)
		require.Equal(t, author1.Name, authorResult.Name)

		var bookResult book
		err = db.Model(&book{}).Find(&bookResult, "name = ?", "Math").Error
		require.NoError(t, err)
		require.Equal(t, book1.Name, bookResult.Name)

		var aBResult authorBook
		err = db.Model(&authorBook{}).Find(&aBResult, "author_id = ?", authorResult.ID).Error
		require.NoError(t, err)
		require.Equal(t, *authorResult.ID, *aBResult.AuthorID)
		require.Equal(t, *bookResult.ID, *aBResult.BookID)
	})

	t.Run("AcrossRepoV2TxFailed_ThenRollback", func(t *testing.T) {
		// reset db after test
		defer func(db *gormv2.DB) {
			err := resetDBV2(db)
			require.NoError(t, err)
		}(dbv2)

		// start TxManager
		txManager := txmanager.StartTxManagerGormV2(dbv2)
		book2 := book{
			Name: stringPointer("Biology"),
		}

		// save author1 data first, to trigger error on below tx
		err := dbv2.Create(&author1).Error
		require.NoError(t, err)

		// doing transaction, add author & book, and then link the author and book
		transaction := func(ctx context.Context) error {
			// saved book2
			bookSaved, err := repoBook.AddBookV2(ctx, book2)

			if err != nil {
				return err
			}

			// saving author1 again will trigger error unique name, because we set on author name must be unique
			authorSaved, err := repoAuthor.AddAuthorV2(ctx, author1)

			if err != nil {
				return err
			}

			_, err = repo.LinkAuthorBookV2(ctx, *authorSaved, *bookSaved)

			if err != nil {
				return err
			}

			return nil
		}

		err = txManager.WithTransaction(context.Background(), transaction)
		require.Error(t, err)

		// validate data
		var bookResult book
		err = dbv2.Model(&book{}).First(&bookResult, "name = ?", "Biology").Error
		require.Error(t, err)
		require.Nil(t, bookResult.ID)
	})

	t.Run("PanicV2_ThenRollback", func(t *testing.T) {
		// reset db after test
		defer func(db *gormv2.DB) {
			err := resetDBV2(db)
			require.NoError(t, err)
		}(dbv2)

		// start TxManager
		txManager := txmanager.StartTxManagerGormV2(dbv2)
		book2 := book{
			Name: stringPointer("Biology"),
		}

		// doing transaction, add author & book, and then link the author and book
		transaction := func(ctx context.Context) error {
			// saved book2
			bookSaved, err := repoBook.AddBookV2(ctx, book2)

			if err != nil {
				return err
			}

			// test have nil pointer value to book
			// when get the value from test, it will trigger panic nil pointer
			var test *book
			failing := *test

			failing.Name = stringPointer("test_panic")

			authorSaved, err := repoAuthor.AddAuthorV2(ctx, author1)

			if err != nil {
				return err
			}

			_, err = repo.LinkAuthorBookV2(ctx, *authorSaved, *bookSaved)

			if err != nil {
				return err
			}

			return nil
		}

		err := txManager.WithTransaction(context.Background(), transaction)
		require.Error(t, err)

		// validate data
		var bookResult book
		err = dbv2.Model(&book{}).First(&bookResult, "name = ?", "Biology").Error
		require.Error(t, err)
		require.Nil(t, bookResult.ID)
	})

}
