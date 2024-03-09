package database

import (
	"database/sql"
)

// build interface for mockDB
type Store interface {
	Querier
}

// real implement of store interface
type SQLStore struct {
	db *sql.DB
	*Queries
}

// instead of using New(db) to directly use queries created by sqlc,
// NewStore(db) can store custom queries.
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,      //db connection
		Queries: New(db), // queries created by sqlc
	}
}

// execute database transaction
// func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
// 	tx, err := store.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return err
// 	}

// 	q := New(tx) // pass tx as parameter instead of db connection
// 	err = fn(q)
// 	if err != nil {
// 		if rbErr := tx.Rollback(); rbErr != nil {
// 			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
// 		}
// 		return err
// 	}

// 	return tx.Commit()

// }
