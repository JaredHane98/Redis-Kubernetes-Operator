package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"
)

type RedisClient struct {
	Client      *redis.Client
	Password    string
	Username    string
	DialTimeout time.Duration
	MaxRetries  int
	DB          int
	Lock        sync.RWMutex
}

const (
	ConnectionAttempts = 3
)

type Employee struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Field     string    `json:"field"`
	StartTime string    `json:"start_time"`
	DOB       string    `json:"dob"`
	Salary    int       `json:"salary"`
}

// caller should have either read/write lock before calling
func (r *RedisClient) IsConnected(ctx context.Context) bool {
	result, err := r.Client.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("failed to ping redis: %v", err)
		return false
	}
	return result == "PONG"
}

func (r *RedisClient) IsConnectedSafe(ctx context.Context, timeout time.Duration) bool {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Lock.RLock()
	defer r.Lock.RUnlock()
	result, err := r.Client.Ping(timeoutCtx).Result()
	if err != nil {
		log.Printf("failed to ping redis: %v", err)
		return false
	}
	return result == "PONG"
}

func (r *RedisClient) CreateRedisClient(tlsConfig *tls.Config, host, port string) error {
	r.Client = redis.NewClient(&redis.Options{
		Addr:                  host + ":" + port,
		Password:              r.Password,
		Username:              r.Username,
		TLSConfig:             tlsConfig,
		DialTimeout:           r.DialTimeout,
		WriteTimeout:          time.Duration(0),
		ReadTimeout:           time.Duration(0),
		MaxRetries:            r.MaxRetries,
		ContextTimeoutEnabled: true,
		DB:                    r.DB,
		PoolSize:              50,
	})

	result, err := r.Client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	if result != "PONG" {
		return fmt.Errorf("failed to ping redis: %s", result)
	}
	return nil
}

func (h *RedisClient) ReadRedis(ctx context.Context, timeout time.Duration, ID string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	h.Lock.RLock()
	defer h.Lock.RUnlock()
	return h.Client.Get(timeoutCtx, ID).Result()
}

func (h *RedisClient) DeleteRedis(ctx context.Context, timeout time.Duration, ID string) (int64, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	h.Lock.RLock()
	defer h.Lock.RUnlock()
	return h.Client.Del(timeoutCtx, ID).Result()
}

func (h *RedisClient) WriteRedis(ctx context.Context, timeout time.Duration, ID string, data []byte, expiration time.Duration) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	h.Lock.RLock()
	defer h.Lock.RUnlock()
	return h.Client.Set(timeoutCtx, ID, data, expiration).Result()
}

type SentinelClient struct {
	Client      *redis.SentinelClient
	Password    string
	Username    string
	MasterName  string
	ServiceName string
	ServicePort string
	DialTimeout time.Duration
	MaxRetries  int
	DB          int
}

func (r *SentinelClient) GetMasterAddrByName(ctx context.Context) ([]string, error) {
	return r.Client.GetMasterAddrByName(ctx, r.MasterName).Result()
}

func (r *SentinelClient) IsConnected(ctx context.Context) bool {
	result, err := r.Client.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("failed to ping redis: %v", err)
		return false
	}
	return result == "PONG"
}

func (r *SentinelClient) UpdateSentinelClient(tlsConfig *tls.Config) error {

	r.Client = redis.NewSentinelClient(&redis.Options{
		Addr:                  r.ServiceName + ":" + r.ServicePort,
		Password:              r.Password,
		Username:              r.Username,
		TLSConfig:             tlsConfig,
		DialTimeout:           r.DialTimeout,
		WriteTimeout:          time.Duration(0), // need to fix
		ReadTimeout:           time.Duration(0), // need to fix
		MaxRetries:            r.MaxRetries,
		ContextTimeoutEnabled: true,
		DB:                    r.DB,
	})

	result, err := r.Client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	if result != "PONG" {
		return fmt.Errorf("failed to ping redis: %s", result)
	}
	return nil
}

type RedisHandle struct {
	TLSConfig      *tls.Config
	RedisClient    *RedisClient
	SentinelClient *SentinelClient
	Updating       int32
	Cond           *sync.Cond
}

func (h *RedisHandle) ReadinessCheck(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func (h *RedisHandle) LivenessCheck(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func (r *RedisHandle) UpdateConnection(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&r.Updating, 0, 1) {
		r.Cond.L.Lock()
		defer r.Cond.L.Unlock()

		for atomic.LoadInt32(&r.Updating) != 0 {
			r.Cond.Wait()
		}
		return nil
	}

	r.RedisClient.Lock.Lock()

	defer func() {
		atomic.StoreInt32(&r.Updating, 0)
		r.Cond.Broadcast() // wake the threads waiting
		r.RedisClient.Lock.Unlock()
	}()

	if !r.SentinelClient.IsConnected(ctx) {
		log.Printf("Updating Sentinel Client")
		if err := r.SentinelClient.Client.Close(); err != nil {
			return err
		}
		if err := r.SentinelClient.UpdateSentinelClient(r.TLSConfig); err != nil {
			return err
		}
	}

	masterAddr, err := r.SentinelClient.GetMasterAddrByName(ctx)
	if err != nil {
		return err
	}
	if len(masterAddr) != 2 {
		return fmt.Errorf("invalid master address: %v", masterAddr)
	}

	log.Printf("Updating master address %s:%s\n", masterAddr[0], masterAddr[1])

	if err := r.RedisClient.Client.Close(); err != nil {
		return err
	}
	if err := r.RedisClient.CreateRedisClient(r.TLSConfig, masterAddr[0], masterAddr[1]); err != nil {
		return err
	}

	return nil
}

func (h *RedisHandle) Add(c *fiber.Ctx) error {

	ctx := context.Background()

	body := c.Body()
	employeeID := gjson.Get(string(body), "id")

	for i := 0; i < ConnectionAttempts; i++ {

		result, err := h.RedisClient.WriteRedis(ctx, time.Second*1, employeeID.String(), body, 0)
		if err == nil {
			return c.Status(fiber.StatusOK).JSON(result)
		}

		if !h.RedisClient.IsConnectedSafe(ctx, time.Second*1) {
			if err := h.UpdateConnection(ctx); err != nil {
				log.Printf("an error occurred while updating the connection: %v", err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
		}
	}
	return c.SendStatus(fiber.StatusRequestTimeout)
}

func (h *RedisHandle) Get(c *fiber.Ctx) error {

	ctx := context.Background()

	employeeID := c.Params("id")
	if employeeID == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	for i := 0; i < ConnectionAttempts; i++ {

		result, err := h.RedisClient.ReadRedis(context.Background(), time.Second*1, employeeID)

		switch err {
		case redis.Nil:
			return c.SendStatus(fiber.StatusNotFound)
		case nil:
			return c.Status(fiber.StatusOK).JSON(result)
		default:
			if !h.RedisClient.IsConnectedSafe(ctx, time.Second*1) {
				if err := h.UpdateConnection(ctx); err != nil {
					log.Printf("an error occurred while updating the connection: %v", err)
					return c.SendStatus(fiber.StatusInternalServerError)
				}
			}
		}
	}
	return c.SendStatus(fiber.StatusRequestTimeout)
}

func (h *RedisHandle) Delete(c *fiber.Ctx) error {

	ctx := context.Background()

	employeeID := c.Params("id")
	if employeeID == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	for i := 0; i < ConnectionAttempts; i++ {
		result, err := h.RedisClient.DeleteRedis(context.Background(), time.Second*1, employeeID)
		switch err {
		case redis.Nil:
			return c.SendStatus(fiber.StatusNotFound)
		case nil:
			return c.Status(fiber.StatusOK).JSON(result)
		default:
			if !h.RedisClient.IsConnectedSafe(ctx, time.Second*1) {
				if err := h.UpdateConnection(ctx); err != nil {
					log.Printf("an error occurred while updating the connection: %v", err)
					return c.SendStatus(fiber.StatusInternalServerError)
				}
			}
		}
	}
	return c.SendStatus(fiber.StatusRequestTimeout)
}

func GetSentinelHandle(config *tls.Config) *SentinelClient {
	sentinel := &SentinelClient{}
	sentinel.MasterName = "mymaster"
	sentinel.DB = 0
	sentinel.DialTimeout = time.Second * 1
	sentinel.MaxRetries = 3
	sentinel.Password = "supersecretpasswordnobodywillguess"
	sentinel.Username = ""
	sentinel.ServicePort = "26379"
	sentinel.ServiceName = "redissentinel-service.redis-database.svc.cluster.local"
	if err := sentinel.UpdateSentinelClient(config); err != nil {
		log.Fatalf("failed to update sentinel client: %v", err)
	}
	return sentinel
}

func GetRedisHandle(config *tls.Config, host, port string) *RedisClient {
	redis := &RedisClient{}
	redis.DB = 0
	redis.DialTimeout = time.Second * 1
	redis.MaxRetries = 3
	redis.Password = "supersecretpasswordnobodywillguess"
	if err := redis.CreateRedisClient(config, host, port); err != nil {
		log.Fatalf("failed to create redis client: %v", err)
	}
	return redis
}

func NewRedisHandleMust() *RedisHandle {
	handle := &RedisHandle{
		Cond: sync.NewCond(&sync.Mutex{}),
	}

	log.Printf("Loading TLS Certificate")
	cer, err := tls.LoadX509KeyPair("tls/tls.crt", "tls/tls.key")
	if err != nil {
		log.Fatalf("failed to read tls certicate and/or key: %v", err)
	}

	caCert, err := os.ReadFile("tls/ca.crt")
	if err != nil {
		log.Fatalf("failed to read ca certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatal("failed to append ca certificate")
	}

	handle.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cer},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	handle.SentinelClient = GetSentinelHandle(handle.TLSConfig)

	addr, err := handle.SentinelClient.GetMasterAddrByName(context.TODO())
	if err != nil {
		log.Fatalf("failed to get master address: %v", err)
	}

	handle.RedisClient = GetRedisHandle(handle.TLSConfig, addr[0], addr[1])

	log.Printf("Successfully connected to redis")

	return handle
}

func main() {

	log.Printf("Starting Redis Worker")

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := fiber.New()

	redisHandle := NewRedisHandleMust()

	app.Get("/liveness", redisHandle.LivenessCheck)
	app.Get("/readiness", redisHandle.ReadinessCheck)
	app.Get("/employee/:id", redisHandle.Get)
	app.Delete("/employee/:id", redisHandle.Delete)
	app.Post("/employee", redisHandle.Add)

	app.All("*", func(c *fiber.Ctx) error {
		errorMessage := fmt.Sprintf("Route '%s' does not exist", c.OriginalURL())

		return c.Status(fiber.StatusNotFound).JSON(&fiber.Map{
			"status":  "fail",
			"message": errorMessage,
		})
	})

	log.Fatal(app.Listen(":8080"))
}
