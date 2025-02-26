package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"social-media-backend/config"
	"social-media-backend/models"
)

func GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data user dari context"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateProfile(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	var input struct {
		Fullname     string `form:"fullname" json:"fullname"`
		Username     string `form:"username" json:"username"`
		Email        string `form:"email" json:"email"`
		JenisKelamin string `form:"jenis_kelamin" json:"jenis_kelamin"`
		TanggalLahir string `form:"tanggal_lahir" json:"tanggal_lahir"` // format YYYY-MM-DD
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var parsedTanggal time.Time
	if input.TanggalLahir != "" {
		t, err := time.Parse("2006-01-02", input.TanggalLahir)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal lahir tidak valid, gunakan YYYY-MM-DD"})
			return
		}
		parsedTanggal = t
	}
	photoPath := currentUser.PhotoProfile
	file, err := c.FormFile("photo_profile")
	if err == nil {
		filename := time.Now().Format("20060102150405_") + file.Filename
		savePath := "public/uploads/" + filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
			return
		}
		photoPath = savePath
	}
	updatedData := models.User{
		Fullname:     input.Fullname,
		Username:     input.Username,
		Email:        input.Email,
		JenisKelamin: input.JenisKelamin,
		PhotoProfile: photoPath,
	}
	if !parsedTanggal.IsZero() {
		updatedData.TanggalLahir = &parsedTanggal
	}
	if err := config.DB.Model(&currentUser).Updates(updatedData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": currentUser})
}

func ChangePassword(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(input.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password lama salah"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 14)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal meng-hash password baru"})
		return
	}
	if err := config.DB.Model(&currentUser).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}

func DeactivateAccount(c *gin.Context) {
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	currentUser := currentUserInterface.(models.User)
	if err := config.DB.Delete(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Akun berhasil dinonaktifkan"})
}

func GetAllUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
};