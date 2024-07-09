/**
 * BTCGO
 *
 * Modulo : Database
 */

package core

import (
	"btcgo/cmd/utils"
	"log"
	"os"
	"path/filepath"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

type dbase struct {
	dbConn  *badger.DB
	DBName  string
	isStart bool
}

// Criar instancia
func NewDatabase() *dbase {
	return &dbase{
		isStart: false,
	}
}

// Start
func (db *dbase) Start(carteira string) {
	rootDir, _ := utils.GetPath()
	op := badger.DefaultOptions(filepath.Join(rootDir, "db", carteira+".db"))
	op.Logger = nil
	dbo, err := badger.Open(op)
	if err != nil {
		log.Fatal(err)
	}
	db.dbConn = dbo
	db.isStart = true
}

// Add Key to DB
func (db *dbase) InsertKey(key string) error {
	err := db.dbConn.Update(func(tx *badger.Txn) error {
		err := tx.Set([]byte(key), []byte(time.Now().Format("2006-01-02 15:04:05")))
		return err
	})
	return err
}

// Verify if Exist key
func (db *dbase) ExistKey(key string) bool {
	err := db.dbConn.View(func(tx *badger.Txn) error {
		_, err := tx.Get([]byte(key))
		return err
	})
	return err == nil
}

// Delete Db
func (db *dbase) Remove(carteira string) error {
	rootDir, _ := utils.GetPath()
	return os.RemoveAll(filepath.Join(rootDir, "db", carteira+".db"))
}

// Stop
func (db *dbase) Stop() {
	if db.isStart {
		db.dbConn.Close()
	}
}
