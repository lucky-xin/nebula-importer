package types

import "k8s.io/utils/env"

type TaskStatus int

var (
	Cipher = env.GetString("DATASOURCE_CIPHER", "6b6579736f6d6574616c6b6579736f6d")
)

var taskStatusMap = map[TaskStatus]string{
	Finished:   "Success",
	Stopped:    "Stopped",
	Scheduling: "Scheduling",
	Running:    "Running",
	NotExisted: "NotExisted",
	Aborted:    "Failed",
	Draft:      "Draft",
}

var taskStatusRevMap = map[string]TaskStatus{
	"finished":   Finished,
	"stopped":    Stopped,
	"running":    Running,
	"scheduling": Scheduling,
	"notExisted": NotExisted,
	"aborted":    Aborted,
	"draft":      Draft,
}

/*
the task in memory (map) has 2 status: processing, aborted;
and the task in local sql has 2 status: finished, stoped;
*/
const (
	StatusUnknown TaskStatus = iota
	Finished
	Stopped
	Scheduling
	Running
	NotExisted
	Aborted
	Draft
)

type (
	CreateImportTaskReq struct {
		Id     *string             `json:"id,optional,omitempty" yaml:"id,omitempty"`
		Name   string              `json:"name" validate:"required" yaml:"name"`
		Tag    string              `json:"tag,omitempty,optional" yaml:"tag,omitempty"`
		Config *ImportTaskV2Config `json:"config" validate:"required" yaml:"config"`
	}
	CreateImportTaskData struct {
		Id string `json:"id"`
	}
	RestartImportTaskReq struct {
		Id *string `path:"id" validate:"required"`
	}
	GetImportTaskReq struct {
		Id string `path:"id" validate:"required"`
	}
	GetImportTaskData struct {
		Id            string          `json:"id"`
		Name          string          `json:"name"`
		User          string          `json:"user"`
		Address       string          `json:"address"`
		ImportAddress []string        `json:"importAddress"`
		Space         string          `json:"space"`
		Status        string          `json:"status"`
		Message       string          `json:"message"`
		CreateTime    int64           `json:"createTime"`
		UpdateTime    int64           `json:"updateTime"`
		Stats         ImportTaskStats `json:"stats"`
		RawConfig     string          `json:"rawConfig"`
	}
	ImportTaskStats struct {
		Processed       int64 `json:"processed"`
		Total           int64 `json:"total"`
		FailedRecords   int64 `json:"failedRecords"`
		TotalRecords    int64 `json:"totalRecords"`
		FailedRequest   int64 `json:"failedRequest"`
		TotalRequest    int64 `json:"totalRequest"`
		TotalLatency    int64 `json:"totalLatency"`
		TotalRespTime   int64 `json:"totalRespTime"`
		FailedProcessed int64 `json:"failedProcessed"`
		TotalProcessed  int64 `json:"totalProcessed"`
	}
	GetManyImportTaskReq struct {
		Page     int    `form:"page,default=1"`
		PageSize int    `form:"pageSize,default=20"`
		Space    string `form:"space,optional"`
		Status   string `form:"status,optional"`
	}
	GetManyImportTaskData struct {
		Total int64               `json:"total"`
		List  []GetImportTaskData `json:"list"`
	}
	GetManyImportTaskLogReq struct {
		Id string `path:"id" validate:"required"`
	}
	DeleteImportTaskReq struct {
		Id string `path:"id"`
	}
	StopImportTaskReq struct {
		Id string `path:"id"`
	}
	ImportTaskCSV struct {
		WithHeader *bool   `json:"withHeader,optional"`
		LazyQuotes *bool   `json:"lazyQuotes,optional"`
		Delimiter  *string `json:"delimiter,optional"`
	}
)

func NewTaskStatus(status string) TaskStatus {
	if v, ok := taskStatusRevMap[status]; ok {
		return v
	}
	return StatusUnknown
}

func (status TaskStatus) String() string {
	if v, ok := taskStatusMap[status]; ok {
		return v
	}
	return "statusUnknown"
}
