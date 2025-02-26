package config

import (
    "log"

    "social-media-backend/models"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
    // Ganti dengan DSN sesuai konfigurasi database kamu.
    dsn := "root:yGkRgDaPonzbtJhrADzRiysVojCLLuTW@tcp(mysql.railway.internal:3306)/railway?charset=utf8mb4&parseTime=True&loc=Local"
    database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Gagal terhubung ke database:", err)
    }

    // AutoMigrate model-model
    err = database.AutoMigrate(
        &models.User{},
        &models.Follow{},
        &models.Feed{},
        &models.Comment{},
        &models.Reaction{},
        &models.Chatroom{},
        &models.Message{},
    )
    if err != nil {
        log.Fatal("AutoMigrate error:", err)
    }

    DB = database
}
