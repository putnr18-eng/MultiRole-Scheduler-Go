package redis

import (
	"context"
	"encoding/json"
	"time"
	"os"

	goredis "github.com/redis/go-redis/v9"
)

var (
	Client *goredis.Client
	Ctx    = context.Background()
)

func Connect() error {
	// Cek apakah variabel REDIS_URL ada di environment (dari Railway)
	redisAddr := os.Getenv("REDIS_URL")
	
	if redisAddr != "" {
		// Jika dijalankan di Railway, gunakan URL lengkap dari REDIS_URL
		opt, err := goredis.ParseURL(redisAddr)
		if err != nil {
			return err
		}
		Client = goredis.NewClient(opt)
	} else {
		// Jika dijalankan di laptop (lokal), gunakan localhost
		Client = goredis.NewClient(&goredis.Options{
			Addr: "127.0.0.1:6379",
			DB:   0,
		})
	}

	_, err := Client.Ping(Ctx).Result()
	return err
}

func RateLimit(key string) (bool, error) {
	count, err := Client.Incr(Ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		Client.Expire(Ctx, key, 300*time.Second)
	}
	if count > 5 {
		return false, nil
	}

	return true, nil
}

// SetC digunakan untuk menyimpan data ke Redis Cache dengan masa kedaluwarsa (duration)
func Set(key string, value interface{}, duration time.Duration) error {
	// Ubah data struct/interface menjadi bentuk JSON string/bytes agar bisa disimpan di Redis
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Simpan menggunakan Client global yang sudah terhubung
	err = Client.Set(Ctx, key, jsonBytes, duration).Err()
	return err
}

// GetC opsional tambahan: Untuk mengambil data cache kembali dari Redis
func Get(key string, dest interface{}) error {
	val, err := Client.Get(Ctx, key).Bytes()
	if err != nil {
		return err // Bisa dicek apakah redis.Nil jika cache tidak ditemukan
	}

	// Unmarshal kembali ke bentuk struct aslinya
	return json.Unmarshal(val, dest)
}
