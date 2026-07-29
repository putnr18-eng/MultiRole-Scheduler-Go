package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

func Connect() error {
	
	redisURL := os.Getenv("REDIS_URL") 
	
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return err
	}

	Client = redis.NewClient(opt)

	_, err = Client.Ping(Ctx).Result()
	return err
}

func RateLimit(key string) (bool, error) {

	if Client == nil {
        fmt.Println("Warning: Redis Client belum terkoneksi, RateLimit dilewati.")
        return true, nil 
    }
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

	if Client == nil {
		fmt.Println("Warning: Redis Client belum terkoneksi, Set dilewati.")
		return nil
	}
	
	j,err:= json.Marshal(value)
	if err!=nil{
		return err
	}
	err= Client.Set(Ctx,key,j,duration ).Err()
	if err != nil {
		// Cukup cetak error atau log, jangan di-return ke fungsi utama
		fmt.Println("Warning: Gagal menyimpan ke Redis:", err)
	}
	return nil
}

func Get(key string)(string,error){
	if Client == nil {
		return "", fmt.Errorf("redis client is nil")
	}

	a,err :=Client.Get(Ctx,key).Result()
	if err!=nil{
		return "", err
	}

	return a,nil
}

func Del(key string)error{
	if Client == nil {
		fmt.Println("Warning: Redis Client belum terkoneksi, Set dilewati.")
		return nil
	}
	err:= Client.Del(Ctx,key).Err()
	if err != nil {
		fmt.Println("Warning: Gagal menghapus dari Redis:", err)
	}
	return nil
}
