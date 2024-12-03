package types

import (
	configv3 "github.com/lucky-xin/nebula-importer/pkg/config/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log"
	"net/url"
	"strconv"
)

type (
	TaskDir struct {
		UploadDir      string `json:"uploadDir,default=./data/upload/"`
		TasksDir       string `json:"tasksDir,default=./data/tasks"`
		CompressionDIR string `json:"compressionDIR,default=./data/compression"`
	}

	DB struct {
		LogLevel                  int                    `json:"logLevel,default=1"`
		IgnoreRecordNotFoundError bool                   `json:"ignoreRecordNotFoundError,default=true"`
		AutoMigrate               bool                   `json:"autoMigrate,default=true"`
		Type                      string                 `json:"type,default=sqlite3"`
		Host                      string                 `json:"host,optional"`
		Name                      string                 `json:"name,optional"`
		User                      string                 `json:"user,optional"`
		Password                  string                 `json:"password,optional"`
		SqliteDbFilePath          string                 `json:"sqliteDbFilePath,default=./data/tasks.db"`
		MaxOpenConns              int                    `json:"maxOpenConns,default=30"`
		MaxIdleConns              int                    `json:"maxIdleConns,default=10"`
		Migrates                  map[string]interface{} `json:"-"`
	}

	Redis struct {
		// The network type, either tcp or unix.
		// Default is tcp.
		Network string `json:"network,default=tcp"`
		URL     string `json:"url,default=redis://:@localhost:6379/0"`

		// ClientName will execute the `CLIENT SETNAME ClientName` command for each conn.
		ClientName string `json:"clientName,default=nebula-studio-gac"`

		// Protocol 2 or 3. Use the version to negotiate RESP version with redis-server.
		// Default is 3.
		Protocol int `json:"protocol,default=3"`

		// Maximum number of retries before giving up.
		// Default is 3 retries; -1 (not 0) disables retries.
		MaxRetries int `json:"maxRetries,default=3"`

		// Base number of socket connections.
		// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
		// If there is not enough connections in the pool, new connections will be allocated in excess of PoolSize,
		// you can limit it through MaxActiveConns
		PoolSize int `json:"poolSize,default=100"`
		// Minimum number of idle connections which is useful when establishing
		// new connection is slow.
		// Default is 0. the idle connections are not closed by default.
		MinIdleConns int `json:"minIdleConns,default=0"`
		// Maximum number of idle connections.
		// Default is 0. the idle connections are not closed by default.
		MaxIdleConns int `json:"maxIdleConns,default=0"`
	}

	Nebula struct {
		Address         string `json:"address" validate:"required"`
		User            string `json:"user" validate:"required"`
		Password        string `json:"password" validate:"required"`
		MaxConnPoolSize int    `json:"maxConnPoolSize,omitempty,optional,default=100"`
		MinConnPoolSize int    `json:"minConnPoolSize,omitempty,optional,default=10"`
		IdleTimeSeconds int    `json:"idleTimeSeconds,omitempty,optional,default=28800"`
		TimeOutSeconds  int    `json:"timeOutSeconds,omitempty,optional,default=10"`
		RiskNGQLRegexp  string `json:"riskNGQLRegexp,omitempty,optional"`
	}

	TaskConfig struct {
		Dir    *TaskDir `json:"dir,optional"`
		Nebula *Nebula  `json:"nebula" validate:"required"`
		DB     *DB      `json:"db" validate:"required"`
		Redis  *Redis   `json:"redis" validate:"required"`

		gormDB *gorm.DB
		rli    *redis.Client
	}

	Log struct {
		Level   *string    `json:"level,omitempty,optional,default=error" yaml:"level,omitempty"`
		Console *bool      `json:"console,omitempty,optional,default=true" yaml:"console,omitempty"`
		Files   []string   `json:"files,omitempty,optional" yaml:"files,omitempty"`
		Fields  []LogField `json:"fields,omitempty,optional" yaml:"fields,omitempty"`
	}

	LogField struct {
		Key   string      `json:"key" yaml:"key,omitempty"`
		Value interface{} `json:"value" yaml:"value,omitempty"`
	}

	Client struct {
		Version                  string  `json:"version,omitempty" validate:"required" yaml:"version"`
		Address                  string  `json:"address,omitempty" validate:"required" yaml:"address"`
		User                     string  `json:"user,omitempty" validate:"required" yaml:"user"`
		Password                 string  `json:"password,omitempty" validate:"required" yaml:"password"`
		ConcurrencyPerAddress    int     `json:"concurrencyPerAddress,optional" yaml:"concurrencyPerAddress,omitempty"`
		ReconnectInitialInterval *string `json:"reconnectInitialInterval,optional,omitempty" yaml:"reconnectInitialInterval,omitempty"`
		Retry                    int     `json:"retry,optional" yaml:"retry,omitempty"`
		RetryInitialInterval     *string `json:"retryInitialInterval,optional,omitempty" yaml:"retryInitialInterval,omitempty"`
	}

	Manager struct {
		SpaceName           string  `json:"spaceName,omitempty" validate:"required" yaml:"spaceName"`
		Batch               int     `json:"batch,omitempty,optional,default=200" yaml:"batch,omitempty"`
		ReaderConcurrency   int     `json:"readerConcurrency,omitempty,optional,default=10" yaml:"readerConcurrency,omitempty"`
		ImporterConcurrency int     `json:"importerConcurrency,omitempty,optional,default=10" yaml:"importerConcurrency,omitempty"`
		StatsInterval       *string `json:"statsInterval,omitempty,optional,default=10s" yaml:"statsInterval,omitempty"`
	}

	ImportTaskV2Config struct {
		configv3.Config `yaml:",inline" json:",inline"`
		Cron            string `json:"cron,omitempty,optional" yaml:"cron,omitempty"`
	}
)

func (tc *TaskConfig) CreateDb() (gormDB *gorm.DB, err error) {
	gormDB = tc.gormDB
	if gormDB != nil {
		return
	}
	openDB, err := OpenDB(tc.DB)
	if err != nil {
		return
	}
	tc.gormDB = openDB
	return
}

func (tc *TaskConfig) CreateRedis() *redis.Client {
	if tc.rli != nil {
		return tc.rli
	}
	u, err := url.Parse(tc.Redis.URL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	password, _ := u.User.Password()
	path := u.Path
	// 假设path为"/0"，提取数据库编号
	dbNumber := 0

	if len(path) > 1 {
		var err error
		dbNumber, err = strconv.Atoi(path[1:])
		if err != nil {
			log.Fatalf("Failed to parse database number: %v", err)
		}

		tc.rli = redis.NewClient(&redis.Options{
			Addr:         u.Host,
			Username:     u.User.Username(),
			Password:     password,
			DB:           dbNumber,
			Protocol:     tc.Redis.Protocol,
			Network:      tc.Redis.Network,
			ClientName:   tc.Redis.ClientName,
			MaxRetries:   tc.Redis.MaxRetries,
			PoolSize:     tc.Redis.PoolSize,
			MinIdleConns: tc.Redis.MinIdleConns,
			MaxIdleConns: tc.Redis.MaxIdleConns,
		})
	}
	return tc.rli
}
