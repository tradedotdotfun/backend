package db

import (
	"tradedotdotfun-backend/common/config"
	"tradedotdotfun-backend/common/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() {
	conn, err := gorm.Open(sqlite.Open(config.SQLITE_DB_PATH), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	db = conn

	conn.AutoMigrate(&model.ChartData{})
	conn.AutoMigrate(&model.Position{})
	conn.AutoMigrate(&model.Account{})
	conn.AutoMigrate(&model.AddressName{})
	conn.AutoMigrate(&model.KeyValueStore{})
	conn.AutoMigrate(&model.LeaderBoard{})
	conn.AutoMigrate(&model.DepositEvent{})
}

func GetConnection() *gorm.DB {
	return db
}
