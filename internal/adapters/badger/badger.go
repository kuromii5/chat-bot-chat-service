package badger

import (
	"context"
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
)

// Cache is an in-memory BadgerDB instance used as a tag validation cache.
type Cache struct {
	db *badger.DB
}

// New opens a BadgerDB in-memory store.
func New() (*Cache, error) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open badger: %w", err)
	}
	return &Cache{db: db}, nil
}

// LoadTags writes all tag names into the cache as keys with empty values.
// Call this once at startup after fetching tags from PostgreSQL.
func (c *Cache) LoadTags(_ context.Context, tags []string) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		for _, tag := range tags {
			if err := txn.Set([]byte(tag), []byte{}); err != nil {
				return fmt.Errorf("txn.Set: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("badger.Update: %w", err)
	}
	return nil
}

// AreTagsValid returns true only if every tag exists in the cache.
func (c *Cache) AreTagsValid(_ context.Context, tags []string) bool {
	err := c.db.View(func(txn *badger.Txn) error {
		for _, tag := range tags {
			if _, err := txn.Get([]byte(tag)); err != nil {
				return fmt.Errorf("txn.Get: %w", err)
			}
		}
		return nil
	})
	return err == nil
}

// Close shuts down the BadgerDB instance.
func (c *Cache) Close() error {
	if err := c.db.Close(); err != nil {
		return fmt.Errorf("badger.Close: %w", err)
	}
	return nil
}
