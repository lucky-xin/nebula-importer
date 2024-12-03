package importer

import (
	"encoding/json"
	"github.com/lucky-xin/nebula-importer/pkg/task/db"
	"github.com/lucky-xin/nebula-importer/pkg/task/types"
	"time"

	"github.com/lucky-xin/nebula-importer/pkg/logger"
	"github.com/lucky-xin/nebula-importer/pkg/manager"
)

type Client struct {
	Cfg        *types.ImportTaskV2Config `json:"cfg,omitempty"`
	Logger     logger.Logger             `json:"logger,omitempty"`
	Manager    manager.Manager           `json:"manager,omitempty"`
	HasStarted bool                      `json:"has_started,omitempty"`
}
type Task struct {
	Client   *Client      `json:"client,omitempty"`
	TaskInfo *db.TaskInfo `json:"task_info,omitempty"`
}

func (t *Task) ToImportTaskData() (result *types.GetImportTaskData, err error) {
	importAddress, err := parseImportAddress(t.TaskInfo.ImportAddress)
	if err != nil {
		return
	}
	stats := t.TaskInfo.Stats
	result = &types.GetImportTaskData{
		Id:            t.TaskInfo.BID,
		Status:        t.TaskInfo.TaskStatus,
		Message:       t.TaskInfo.TaskMessage,
		CreateTime:    t.TaskInfo.CreateTime.UnixMilli(),
		UpdateTime:    t.TaskInfo.UpdateTime.UnixMilli(),
		Address:       t.TaskInfo.Address,
		ImportAddress: importAddress,
		User:          t.TaskInfo.User,
		Name:          t.TaskInfo.Name,
		Space:         t.TaskInfo.Space,
		RawConfig:     t.TaskInfo.RawConfig,
		Stats: types.ImportTaskStats{
			Total:           stats.Total,
			Processed:       stats.Processed,
			FailedRecords:   stats.FailedRecords,
			TotalRecords:    stats.TotalRecords,
			TotalRequest:    stats.TotalRequest,
			FailedRequest:   stats.FailedRequest,
			TotalLatency:    int64(stats.TotalLatency),
			TotalRespTime:   int64(stats.TotalRespTime),
			FailedProcessed: stats.FailedProcessed,
			TotalProcessed:  stats.TotalProcessed,
		},
	}
	return
}

func (t *Task) Marshal() (byts []byte, err error) {
	byts, err = json.Marshal(t)
	return
}

func UnmarshalTask(byts []byte) (t *Task, err error) {
	var task = Task{}
	err = json.Unmarshal(byts, &task)
	t = &task
	return
}

func (t *Task) UpdateQueryStats() error {
	if (t.Client == nil) || (t.Client.Manager == nil) {
		return nil
	}
	stats := t.Client.Manager.Stats()
	t.TaskInfo.Stats = db.Stats{
		Processed:       stats.Processed,
		Total:           stats.Total,
		FailedRecords:   stats.FailedRecords,
		TotalRecords:    stats.TotalRecords,
		FailedRequest:   stats.FailedRequest,
		TotalRequest:    stats.TotalRequest,
		TotalLatency:    stats.TotalLatency,
		TotalRespTime:   stats.TotalRespTime,
		FailedProcessed: stats.FailedProcessed,
		TotalProcessed:  stats.TotalProcessed,
	}
	t.TaskInfo.UpdateTime = time.Now()
	return nil
}
