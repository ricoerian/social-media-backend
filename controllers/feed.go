package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"social-media-backend/config"
	"social-media-backend/models"
)

// GetFeeds mengembalikan daftar feeds, beserta data user dan komentar.
func GetFeeds(c *gin.Context) {
	var feeds []models.Feed
	// Preload User, Comments, dan Reactions agar data terkait ikut ter-fetch.
	if err := config.DB.Preload("User").
		Preload("Comments.User").
		Preload("Reactions").
		Order("created_at desc").
		Find(&feeds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feeds": feeds})
}

// FeedInput digunakan untuk validasi data pembuatan dan update feed.
type FeedInput struct {
	Feed string `json:"feed" binding:"required"`
	File string `json:"file"`
}

// CreateFeed memungkinkan user membuat feed baru.
func CreateFeed(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)

	// Gunakan binding yang mendukung form-data, bukan JSON.
	var input struct {
		Feed string `form:"feed" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Proses file upload jika ada
	var filePaths []string
	// Ambil semua file dengan key "file"
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files := form.File["file"]
		for _, file := range files {
			// Buat nama file unik
			filename := time.Now().Format("20060102150405_") + file.Filename
			savePath := "public/uploads/" + filename
			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
				return
			}
			filePaths = append(filePaths, savePath)
		}
	}

	// Gabungkan path file yang diupload menjadi string yang dipisahkan koma
	filesString := strings.Join(filePaths, ",")

	feed := models.Feed{
		Feed:      input.Feed,
		File:      filesString,
		UserID:    currentUser.ID,
		CreatedAt: time.Now(),
	}

	if err := config.DB.Create(&feed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"feed": feed})
}

// UpdateFeed memungkinkan pemilik feed untuk mengedit feed-nya, termasuk mengganti file/foto (multiple file)
func UpdateFeed(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)

	feedIDStr := c.Param("feed_id")
	feedID, err := strconv.ParseUint(feedIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID feed tidak valid"})
		return
	}

	var feed models.Feed
	if err := config.DB.First(&feed, uint(feedID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feed tidak ditemukan"})
		return
	}

	// Pastikan feed dimiliki oleh user yang sedang login.
	if feed.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak untuk mengubah feed ini"})
		return
	}

	// Binding form-data untuk text feed
	var input struct {
		Feed string `form:"feed" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Proses file upload jika ada file baru (bisa lebih dari satu)
	var filePaths []string
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files := form.File["file"]
		for _, file := range files {
			filename := time.Now().Format("20060102150405_") + file.Filename
			savePath := "public/uploads/" + filename
			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
				return
			}
			filePaths = append(filePaths, savePath)
		}
	}

	// Jika ada file baru diupload, update field File, jika tidak, biarkan nilai lama
	if len(filePaths) > 0 {
		feed.File = strings.Join(filePaths, ",")
	}

	feed.Feed = input.Feed
	feed.UpdatedAt = time.Now()

	if err := config.DB.Save(&feed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feed": feed})
}

// DeleteFeed memungkinkan pemilik feed untuk menghapus feed-nya (soft delete)
func DeleteFeed(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)

	feedIDStr := c.Param("feed_id")
	feedID, err := strconv.ParseUint(feedIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID feed tidak valid"})
		return
	}

	var feed models.Feed
	if err := config.DB.First(&feed, uint(feedID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feed tidak ditemukan"})
		return
	}

	if feed.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak untuk menghapus feed ini"})
		return
	}

	if err := config.DB.Delete(&feed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feed berhasil dihapus"})
}

// CommentInput digunakan untuk validasi data pembuatan comment.
type CommentInput struct {
	Comment string `json:"comment" binding:"required"`
	File    string `json:"file"`
}

// CreateComment memungkinkan user menambahkan komentar pada feed.
func CreateComment(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)

	feedIDStr := c.Param("feed_id")
	feedID, err := strconv.ParseUint(feedIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID feed tidak valid"})
		return
	}

	var feed models.Feed
	if err := config.DB.First(&feed, uint(feedID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feed tidak ditemukan"})
		return
	}

	var input CommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := models.Comment{
		Comment:   input.Comment,
		File:      input.File,
		FeedID:    feed.ID,
		UserID:    currentUser.ID,
		CreatedAt: time.Now(),
	}

	if err := config.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Preload data User untuk memasukkan data user yang membuat komentar
	if err := config.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

// UpdateComment memungkinkan pengirim comment untuk mengedit komentarnya.
func UpdateComment(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)

	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID comment tidak valid"})
		return
	}

	var comment models.Comment
	if err := config.DB.First(&comment, uint(commentID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment tidak ditemukan"})
		return
	}

	// Hanya pengirim komentar yang dapat mengedit komentarnya.
	if comment.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak untuk mengedit comment ini"})
		return
	}

	var input struct {
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.Comment = input.Comment
	comment.UpdatedAt = time.Now()

	if err := config.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment berhasil diupdate", "comment": comment})
}

// DeleteComment memungkinkan user menghapus comment yang dibuatnya.
func DeleteComment(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)

	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID comment tidak valid"})
		return
	}

	var comment models.Comment
	if err := config.DB.First(&comment, uint(commentID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment tidak ditemukan"})
		return
	}

	// Hanya pemilik comment yang dapat menghapusnya.
	if comment.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak untuk menghapus comment ini"})
		return
	}

	if err := config.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment berhasil dihapus"})
}

// LikeFeed memungkinkan user memberikan reaksi "like" pada feed.
func LikeFeed(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)

	feedIDStr := c.Param("feed_id")
	feedID, err := strconv.ParseUint(feedIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID feed tidak valid"})
		return
	}

	var feed models.Feed
	if err := config.DB.First(&feed, uint(feedID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feed tidak ditemukan"})
		return
	}

	var reaction models.Reaction
	err = config.DB.Where("feed_id = ? AND user_id = ?", feed.ID, currentUser.ID).First(&reaction).Error
	if err == nil {
		// Jika sudah ada reaksi, periksa tipe reaksi.
		if reaction.Reaction == "like" {
			if err := config.DB.Delete(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus like"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Like dihapus"})
			return
		} else {
			reaction.Reaction = "like"
			if err := config.DB.Save(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate reaksi"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Diupdate menjadi like"})
			return
		}
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newReaction := models.Reaction{
		FeedID:    feed.ID,
		UserID:    currentUser.ID,
		Reaction:  "like",
		CreatedAt: time.Now(),
	}
	if err := config.DB.Create(&newReaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feed dilike"})
}

// DislikeFeed memungkinkan user memberikan reaksi "dislike" pada feed.
func DislikeFeed(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)

	feedIDStr := c.Param("feed_id")
	feedID, err := strconv.ParseUint(feedIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID feed tidak valid"})
		return
	}

	var feed models.Feed
	if err := config.DB.First(&feed, uint(feedID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feed tidak ditemukan"})
		return
	}

	var reaction models.Reaction
	err = config.DB.Where("feed_id = ? AND user_id = ?", feed.ID, currentUser.ID).First(&reaction).Error
	if err == nil {
		if reaction.Reaction == "dislike" {
			if err := config.DB.Delete(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus dislike"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Dislike dihapus"})
			return
		} else {
			reaction.Reaction = "dislike"
			if err := config.DB.Save(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate reaksi"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Diupdate menjadi dislike"})
			return
		}
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newReaction := models.Reaction{
		FeedID:    feed.ID,
		UserID:    currentUser.ID,
		Reaction:  "dislike",
		CreatedAt: time.Now(),
	}
	if err := config.DB.Create(&newReaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feed didislike"})
}
