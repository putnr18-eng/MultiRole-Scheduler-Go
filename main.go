package main

import (
	"context"

	"log"
	"net/http"
	"os"
	"os/signal"
	"github.com/gin-contrib/cors"
	"syscall"
	"time"

	"play/database"
	"play/internal/auth"
	"play/middleware"
	"play/redis"
	"play/user"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)



func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("gagal load env")
	}

	pool, err := database.Connect()
	if err != nil {
		log.Fatalf("gagal konek ke database %v", err)
	}
	defer pool.Close()

	c := cache.New(24*time.Hour, 10*time.Minute)

	auth := &auth.Data{
		DB:    pool,
		Cache: c,
	}

	user := &user.DB{
		Database: pool,
		Cache:    c,
	}
	redis.Connect()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://goojadwal.pages.dev"}, // Sesuaikan dengan frontend production kamu
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/regis", auth.Register)
	r.POST("/login", auth.Login)
	r.POST("/tes", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ok": "ok",
		})
	})

	r.GET("/api/public/jadwal/:id", user.MemberProfile)
	r.GET("/api/booking/jadwal/:username", user.BookingProfile)

	// Endpoint untuk booking dari publik/tamu tanpa login:
	r.POST("/api/booking/:username", user.CreateBooking)

	// --- 2. RUTE YANG BUTUH LOGIN (Masuk ke Group Protected) ---
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{

		protected.GET("/download", func(c *gin.Context) {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		}, user.Profile)

		// Endpoint Logout
		protected.POST("/logout", auth.Logout)

		protected.GET("/user", user.Profile)
		protected.PUT("/user/accept/:id", user.Accept)
		protected.DELETE("/user/jadwal/:id", user.Delete)
		protected.PUT("/user/jadwal/keterangan/:id", user.UpdateKeterangan)

		// --- ENDPOINT BARU GANTI PASSWORD & USERNAME ---
		protected.PUT("/user/change-password", user.UpdatePassword)
		protected.PUT("/user/change-username", user.UpdateUsername)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port, // Menggunakan port dinamis dari cloud atau fallback ke 8080
		Handler: r,

		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server gagal jalan %v", err)
		}
	}()

	log.Printf("server jalan di port %s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("mematikan server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server down gagal shutdown", err)
	}

	log.Print("server berhenti dengan aman")
}
