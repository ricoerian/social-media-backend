package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"social-media-backend/config"
	"social-media-backend/models"
)

func FollowUser(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	targetIDStr := c.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID user tidak valid"})
		return
	}
	var targetUser models.User
	if err := config.DB.First(&targetUser, uint(targetID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	var followingList []models.User
	config.DB.Model(&currentUser).Association("Following").Find(&followingList, "id = ?", targetUser.ID)
	if len(followingList) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Anda sudah mengikuti user ini"})
		return
	}
	if err := config.DB.Model(&currentUser).Association("Following").Append(&targetUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengikuti user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengikuti user"})
}

func UnfollowUser(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	targetIDStr := c.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID user tidak valid"})
		return
	}
	var targetUser models.User
	if err := config.DB.First(&targetUser, uint(targetID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	if err := config.DB.Model(&currentUser).Association("Following").Delete(&targetUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal berhenti mengikuti user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil berhenti mengikuti user"})
}

func GetFollowers(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	var followers []models.User
	if err := config.DB.Model(&currentUser).Association("Followers").Find(&followers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar followers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"followers": followers})
}

func GetFollowing(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	var following []models.User
	if err := config.DB.Model(&currentUser).Association("Following").Find(&following); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar following"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"following": following})
};