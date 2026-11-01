package config
import(
	"os"
	"strconv"
	"time")
 
type Config struct{
	AppDomain string
	HTTPAddr  string
	DatabaseURL string
	RedisAddr string
	LicenseFile string
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	IdleTimeout time.Duration
	ShutdownTimeout time.Duration
}

func Load() Config{
	return Config{
	AppDomain: getEnv("APP_DOMAIN", "localhost"),
	HTTPAddr : getEnv("HTTP_ADDR",":8000"),
	DatabaseURL:     getEnv("DATABASE_URL", "postgres://ims:ims_secret@localhost:5432/ims?sslmode=disable"),
	RedisAddr : getEnv("REDIS_ADDR", "localhost:6379"),
	LicenseFile : getEnv("MINTERN_LICENSE_FILE" , "/etc/mintern/mintern_enterprise.llc"),
	ReadTimeout:     getDurationeEnv("HTTP_READ_TIMEOUT_SEC", 10) * time.Second,
	WriteTimeout:    getDurationeEnv("HTTP_WRITE_TIMEOUT_SEC", 15) * time.Second,
	IdleTimeout:     getDurationeEnv("HTTP_IDLE_TIMEOUT_SEC", 60) * time.Second,
	ShutdownTimeout: getDurationeEnv("HTTP_SHUTDOWN_TIMEOUT_SEC", 10) * time.Second,}

	}
  
func getEnv(key , fallback string)string{
	if val,ok := os.LookupEnv(key); ok && val != ""{
		return val
	}
	return fallback
}

func getDurationeEnv(key string , fallbackseconds int) time.Duration {
	val , ok := os.LookupEnv(key)
	if !ok || val == ""{
		return time.Duration(fallbackseconds)
	}
	parsed,err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		return time.Duration(fallbackseconds)
	}
	return time.Duration(parsed)
}