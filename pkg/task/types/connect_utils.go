package types

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"iter"
	"log"
	"maps"
	"os"
	"strings"
	"time"
)

func ParseDSN(opts *DB) (dsn string, err error) {
	switch opts.Type {
	case "mysql":
		concate := "?"
		if strings.Contains(opts.Name, concate) {
			concate = "&"
		}
		if opts.Host[0] == '/' {
			dsn = fmt.Sprintf("%s:%s@unix(%s)/%s%scharset=utf8mb4&parseTime=true&loc=Local",
				opts.User, opts.Password, opts.Host, opts.Name, concate)
		} else {
			dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s%scharset=utf8mb4&parseTime=true&loc=Local",
				opts.User, opts.Password, opts.Host, opts.Name, concate)
		}

	case "sqlite3":
		dsn = "file:" + opts.SqliteDbFilePath +
			"?cache=shared&mode=rwc&synchronous=off&busy_timeout=10000&journal_mode=wal"

	default:
		return "", errors.Errorf("unrecognized dialect: %s", opts.Type)
	}

	return dsn, nil
}

func OpenDB(opts *DB) (*gorm.DB, error) {
	logx.Info(opts)
	dsn, err := ParseDSN(opts)
	if err != nil {
		return nil, errors.Wrap(err, "parse DSN")
	}

	var dialector gorm.Dialector
	switch opts.Type {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite3":
		dialector = sqlite.Open(dsn)
	}

	cfg := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				IgnoreRecordNotFoundError: opts.IgnoreRecordNotFoundError,
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.LogLevel(opts.LogLevel),
				Colorful:                  false,
				ParameterizedQueries:      false,
			},
		),
	}

	gormDB, err := gorm.Open(dialector, cfg)
	if gormDB != nil && opts.Type == "mysql" {
		gormDB.Set("gorm:table_options", "ENGINE=InnoDB")
	}
	return gormDB, err
}

// InitSQLDB initialize local sql by open sql and create task_infos table
func InitSQLDB(config *DB, gormDB *gorm.DB) {
	if gormDB == nil {
		var err error
		gormDB, err = OpenDB(config)
		if err != nil {
			zap.L().Fatal(fmt.Sprintf("init db fail: %s", err))
			panic(err)
		}
		sqlDB, err := gormDB.DB()
		if err != nil {
			zap.L().Fatal(fmt.Sprintf("init db fail: %v", err))
			panic(err)
		}

		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
		sqlDB.SetConnMaxIdleTime(time.Hour)
	}

	if config.AutoMigrate {
		migrateTables := maps.Keys(config.Migrates)
		err := MigrateAlterBID(gormDB, migrateTables)
		if err != nil {
			zap.L().Fatal(fmt.Sprintf("migrate tables fail: %s", err))
			panic(err)
		}
		values := maps.Values(config.Migrates)
		var dst []interface{}
		for i := range values {
			dst = append(dst, i)
		}
		if len(dst) == 0 {
			return
		}
		err = gormDB.AutoMigrate(dst...)
		if err != nil {
			zap.L().Fatal(fmt.Sprintf("init taskInfo table fail: %s", err))
			panic(err)
		}
	}
}

// MigrateAlterBID Add the `b_id“ field to certain database tables for versions prior to v3.7 (inclusive).
// We need to perform this operation manually, otherwise the `db.AutoMigrate` method will throw an exception
func MigrateAlterBID(db *gorm.DB, tables iter.Seq[string]) error {
	for table := range tables {
		isTableExist := db.Migrator().HasTable(table)
		// if table not exists, skip
		if !isTableExist {
			continue
		}

		isColumnExist, err := ColumnExists(db, table, "b_id")
		if err != nil {
			return err
		}
		// if column exists, skip
		if isColumnExist {
			continue
		}

		err = db.Exec("ALTER TABLE `" + table + "` ADD COLUMN `b_id` CHAR(32) NOT NULL DEFAULT ''").Error
		if err != nil {
			return err
		}
		err = db.Exec("UPDATE `" + table + "` SET `b_id` = `id`").Error
		if err != nil {
			return err
		}
	}
	return nil
}

func ColumnExists(db *gorm.DB, tableName string, columnName string) (bool, error) {
	// mysql | postgres
	query := "SELECT count(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ? AND column_name = ?"

	dbType := db.Dialector.Name()
	// sqlite does not support `INFORMATION_SCHEMA`, so we need to use `PRAGMA table_info`
	if dbType == "sqlite" {
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	}

	rows, err := db.Raw(query, tableName, columnName).Rows()
	if err != nil {
		return false, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logx.Error(err)
		}
	}(rows)

	for rows.Next() {
		if dbType == "sqlite" {
			var cid, pk int
			var name, dataType, dfltValue string
			var notnull bool
			/*
				| cid | name | type         | notnull | dflt_value | pk |
				|-----|------|--------------|---------|------------|----|
				| 0   | id   | INTEGER      | 0       | NULL       | 1  |
				| 1   | b_id | char(32)     | 1       | NULL       | 0  |
				| 2   | name | varchar(255) | 1       | NULL       | 0  |
			*/
			_ = rows.Scan(&cid, &name, &dataType, &notnull, &dfltValue, &pk)
			if name == columnName {
				return true, nil
			}
		} else {
			var count int
			_ = rows.Scan(&count)
			if count > 0 {
				return true, nil
			}
		}
	}

	return false, nil
}
