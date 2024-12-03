package db

import (
	"time"
)

type Stats struct {
	Processed       int64         `gorm:"column:processed;" json:"processed"`
	Total           int64         `gorm:"column:total;" json:"total"`
	FailedRecords   int64         `gorm:"column:failed_records;" json:"failed_records"`
	TotalRecords    int64         `gorm:"column:total_records;" json:"total_records"`
	FailedRequest   int64         `gorm:"column:failed_request;" json:"failed_request"`
	TotalRequest    int64         `gorm:"column:total_request;" json:"total_request"`
	TotalLatency    time.Duration `gorm:"column:total_latency;" json:"total_latency"`
	TotalRespTime   time.Duration `gorm:"column:total_resp_time;" json:"total_resp_time"`
	FailedProcessed int64         `gorm:"column:failed_processed;" json:"failed_processed"`
	TotalProcessed  int64         `gorm:"column:total_processed;" json:"total_processed"`
}

type TaskInfo struct {
	ID            int    `gorm:"column:id;primaryKey;autoIncrement;" json:"id"`
	BID           string `gorm:"column:b_id;not null;type:char(32);uniqueIndex;comment:task id" json:"bid"`
	Address       string `gorm:"column:address;type:varchar(255);" json:"address"`
	Name          string `gorm:"column:name;type:varchar(255);" json:"name"`
	Space         string `gorm:"column:space;type:varchar(255);" json:"space"`
	ImportAddress string `gorm:"column:import_address;" json:"import_address"`
	User          string `gorm:"column:user;" json:"user"`
	TaskStatus    string `gorm:"column:task_status;" json:"task_status"`
	TaskMessage   string `gorm:"column:task_message;" json:"task_message"`
	Stats         Stats  `gorm:"embedded" json:"stats"`
	RawConfig     string `gorm:"column:raw_config;type:mediumtext;" json:"raw_config"`
	Cron          string `gorm:"column:cron;not null;type:varchar(200);default:''" json:"cron"`

	CreateTime time.Time `gorm:"column:create_time;type:datetime;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime;autoUpdateTime" json:"update_time"`
}

// TaskEffect storage for task yaml config and partial task log
type TaskEffect struct {
	ID     int    `gorm:"column:id;primaryKey;autoIncrement;"`
	BID    string `gorm:"column:task_id;not null;type:char(32);uniqueIndex;comment:task id"`
	Log    string `gorm:"column:log;type:mediumtext;comment:partial task log"`
	Config string `gorm:"column:config;type:mediumtext;comment:task config.yaml"`

	CreateTime time.Time `gorm:"column:create_time;type:datetime;autoCreateTime"`
}
