package main

import (
    "os"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "social-media-backend/config"
    "social-media-backend/controllers"
)

func main() {
    config.ConnectDatabase()

    r := gin.Default()

    // Konfigurasi CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"https://social-media-frontend-beryl-mu.vercel.app"}, // Sesuaikan dengan URL frontend kamu
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    // Endpoint autentikasi.
    r.Static("/public", "./public")
    r.POST("/register", controllers.Register)
    r.POST("/login", controllers.Login)

    // Group endpoint yang dilindungi oleh autentikasi.
    authorized := r.Group("/")
    authorized.Use(controllers.AuthMiddleware())
    {
        // Endpoint profile user.
        authorized.GET("/profile", controllers.GetProfile)
        authorized.PUT("/profile", controllers.UpdateProfile)
        authorized.PUT("/profile/password", controllers.ChangePassword)
        authorized.DELETE("/profile", controllers.DeactivateAccount)

        // Endpoint follow.
        authorized.POST("/follow/:id", controllers.FollowUser)
        authorized.DELETE("/follow/:id", controllers.UnfollowUser)
        authorized.GET("/followers", controllers.GetFollowers)
        authorized.GET("/following", controllers.GetFollowing)

        // Endpoint feeds & comments.
		authorized.GET("/feeds", controllers.GetFeeds)
        authorized.POST("/feeds", controllers.CreateFeed)
        authorized.PUT("/feeds/:feed_id", controllers.UpdateFeed)
        authorized.DELETE("/feeds/:feed_id", controllers.DeleteFeed)
        authorized.POST("/feeds/:feed_id/comments", controllers.CreateComment)
        authorized.PUT("/comments/:id", controllers.UpdateComment)
        authorized.DELETE("/comments/:id", controllers.DeleteComment)
        authorized.POST("/feeds/:feed_id/like", controllers.LikeFeed)
        authorized.POST("/feeds/:feed_id/dislike", controllers.DislikeFeed)

        // Endpoint chat.
		authorized.GET("/users", controllers.GetAllUsers)
        authorized.POST("/chatrooms", controllers.CreateChatroom)
        authorized.GET("/chatrooms", controllers.GetChatrooms)
        authorized.GET("/chatrooms/:id/messages", controllers.GetChatroomMessages)
        authorized.POST("/chatrooms/:id/messages", controllers.SendMessage)
        authorized.DELETE("/chatrooms/:id", controllers.DeleteChatroom)
        authorized.DELETE("/messages/:id", controllers.DeleteMessage)
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
};