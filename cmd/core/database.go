/**
 * BTCGO
 *
 * Modulo : Database
 */

package core

import (
	"log"
	"time"

	//"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type dbase struct {
	dbConn *gorm.DB
	dbName string
}

type TestedKeys struct {
	Key      string `gorm:"primaryKey; not null"`
	Carteira string `gorm:"index; not null"`
	DataHora string
}

// Criar instancia
func NewDatabase(fileName string) *dbase {
	db, err := gorm.Open(sqlite.Open(fileName), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		log.Fatal("Cant open database", err)
	}
	err = db.AutoMigrate(&TestedKeys{})
	if err != nil {
		log.Fatal(err)
	}
	return &dbase{
		dbConn: db,
		dbName: fileName,
	}
}

// Add Key to DB
func (db *dbase) InsertKey(carteira, key string) error {
	result := db.dbConn.Create(TestedKeys{
		Key:      key,
		Carteira: carteira,
		DataHora: time.Now().Format("2006-01-02 15:04:05"),
	})
	if result.Error != nil {
		log.Println("INSERT Db error", result.Error)
	}
	return result.Error
}

// Verify if Exist key
func (db *dbase) ExistKey(key string) bool {
	var get TestedKeys
	result := db.dbConn.First(&get, "Key=?", key)
	//log.Println(result.Error)
	return result.Error == nil
}

// Apaga todas as chaves para a carteira
func (db *dbase) Delete(carteira string) error {
	result := db.dbConn.Delete(carteira)
	if result.Error != nil {
		log.Println("DELETE Db error", result.Error)
	}
	return result.Error
}
