package txmanager_test

import (
	"context"
	"github.com/shortlyst-ai/go-txmanager"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTxManager_WithTransaction(t *testing.T) {
	db := getMysql()
	db.AutoMigrate(&author{}, &book{}, &authorBook{})
	defer closeDB(t, db)

	author1 := author{
		Name: stringPointer("john"),
	}
	book1 := book{
		Name: stringPointer("Math"),
	}

	repoAuthor := newAuthorRepository(db)
	repoBook := newBookRepository(db)
	repo := newAuthorBookRepository(db)

	t.Run("AcrossRepoTx_Success", func(t *testing.T) {
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

	t.Run("Panic_ThenRollback", func(t *testing.T) {
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
}