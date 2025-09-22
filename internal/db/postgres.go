package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectWithRetry пытается подключиться attempts раз с паузой wait
func ConnectWithRetry(dsn string, attempts int, wait time.Duration) (*gorm.DB, error) {
	var gormDB *gorm.DB
	var err error
	for i := 0; i < attempts; i++ {
		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			sqlDB, err2 := gormDB.DB()
			if err2 == nil {
				if err3 := sqlDB.Ping(); err3 == nil {
					return gormDB, nil
				} else {
					err = err3
				}
			} else {
				err = err2
			}
		}
		fmt.Printf("DB connect attempt %d/%d failed: %v. retrying in %s...\n", i+1, attempts, err, wait)
		time.Sleep(wait)
	}
	return nil, err
}
