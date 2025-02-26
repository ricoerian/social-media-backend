package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"social-media-backend/config"
	"social-media-backend/models"
)

type ChatroomInput struct {
	IsGroup bool   `json:"is_group"`
	Name    string `json:"name"`     // wajib untuk group chat
	UserIDs []uint `json:"user_ids"` // user yang akan diikutkan (selain current user)
}

func CreateChatroom(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	var input ChatroomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !input.IsGroup && len(input.UserIDs) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Untuk direct chat, harus ada tepat satu user ID"})
		return
	}
	if input.IsGroup && input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama grup harus diisi untuk group chat"})
		return
	}
	chatroom := models.Chatroom{
		IsGroup:   input.IsGroup,
		Name:      input.Name,
		CreatedAt: time.Now(),
		OwnerID:   currentUser.ID, // untuk group chat, set OwnerID
	}
	if err := config.DB.Create(&chatroom).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := config.DB.Model(&chatroom).Association("Users").Append(&currentUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan user ke chatroom"})
		return
	}
	for _, uid := range input.UserIDs {
		var user models.User
		if err := config.DB.First(&user, uid).Error; err != nil {
			continue
		}
		if err := config.DB.Model(&chatroom).Association("Users").Append(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan user ke chatroom"})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"chatroom": chatroom})
}

func GetChatrooms(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	var chatrooms []models.Chatroom
	if err := config.DB.Model(&currentUser).Association("Chatrooms").Find(&chatrooms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil chatrooms"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"chatrooms": chatrooms})
}

func GetChatroomMessages(c *gin.Context) {
	chatroomIDStr := c.Param("id")
	chatroomID, err := strconv.ParseUint(chatroomIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chatroom tidak valid"})
		return
	}
	var chatroom models.Chatroom
	if err := config.DB.First(&chatroom, uint(chatroomID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chatroom tidak ditemukan"})
		return
	}
	var messages []models.Message
	if err := config.DB.Where("chatroom_id = ?", chatroom.ID).
		Preload("User").
		Order("created_at asc").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil pesan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func SendMessage(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	chatroomIDStr := c.Param("id")
	chatroomID, err := strconv.ParseUint(chatroomIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chatroom tidak valid"})
		return
	}
	var chatroom models.Chatroom
	if err := config.DB.First(&chatroom, uint(chatroomID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chatroom tidak ditemukan"})
		return
	}
	messageText := c.PostForm("message")
	if messageText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pesan tidak boleh kosong"})
		return
	}
	filePath := ""
	file, err := c.FormFile("file")
	if err == nil {
		filename := time.Now().Format("20060102150405_") + file.Filename
		savePath := "public/uploads/" + filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
			return
		}
		filePath = savePath
	}
	message := models.Message{
		Message:    messageText,
		File:       filePath,
		ChatroomID: chatroom.ID,
		UserID:     currentUser.ID,
		CreatedAt:  time.Now(),
	}
	if err := config.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim pesan"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": message})
}

func DeleteChatroom(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	chatroomIDStr := c.Param("id")
	chatroomID, err := strconv.ParseUint(chatroomIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chatroom tidak valid"})
		return
	}
	var chatroom models.Chatroom
	if err := config.DB.First(&chatroom, uint(chatroomID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chatroom tidak ditemukan"})
		return
	}
	if chatroom.IsGroup {
		// Untuk group chat, hanya Owner yang boleh menghapus.
		if chatroom.OwnerID != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak untuk menghapus group chat ini"})
			return
		}
	} else {
		// Untuk direct chat, pastikan currentUser adalah peserta.
		var participants []models.User
		if err := config.DB.Model(&chatroom).Association("Users").Find(&participants); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data chatroom"})
			return
		}
		allowed := false
		for _, u := range participants {
			if u.ID == currentUser.ID {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak untuk menghapus chatroom ini"})
			return
		}
	}
	if err := config.DB.Delete(&chatroom).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus chatroom"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chatroom berhasil dihapus"})
}

func DeleteMessage(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pesan tidak valid"})
		return
	}
	var message models.Message
	if err := config.DB.First(&message, uint(messageID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesan tidak ditemukan"})
		return
	}
	if message.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak untuk menghapus pesan ini"})
		return
	}
	if err := config.DB.Delete(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pesan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesan berhasil dihapus"})
};