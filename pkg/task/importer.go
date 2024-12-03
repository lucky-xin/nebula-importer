package importer

import (
	"errors"
	"fmt"
	"github.com/lucky-xin/nebula-importer/pkg/task/ecode"
	"github.com/lucky-xin/nebula-importer/pkg/task/types"
	"regexp"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type ImportResult struct {
	TaskId      string `json:"taskId"`
	TimeCost    string `json:"timeCost"` // Milliseconds
	FailedRows  int64  `json:"failedRows"`
	ErrorResult struct {
		ErrorCode int    `json:"errorCode"`
		ErrorMsg  string `json:"errorMsg"`
	}
}

func StartImport(taskID string) (err error) {
	task, b := GetTaskMgr().GetTask(taskID)
	if !b {
		return errors.New("not found task,id:" + taskID)
	}
	signal := make(chan struct{}, 1)
	if task.TaskInfo.TaskStatus != types.Running.String() {
		task.TaskInfo.TaskStatus = types.Running.String()
		err = GetTaskMgr().UpdateTaskInfoById(task)
		if err != nil {
			logx.Errorf("update task status error: %v", err)
			return
		}
	}
	abort := func() {
		_ = GetTaskMgr().AbortTask(taskID, err.Error())
		signal <- struct{}{}
	}
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = GetTaskMgr().UpdateTaskInfo(taskID)
			case <-signal:
				return
			}
		}
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logx.Errorf("[task import error]: &s, %+v", err)
				_ = GetTaskMgr().AbortTask(taskID, fmt.Sprintf("%v", err))
			}
		}()
		cfg := task.Client.Cfg
		// 最终会调用Manager.Import方法开始导入数据
		if err = cfg.Build(); err != nil {
			logx.Errorf("build error: %v", err)
			abort()
			return
		}
		mgr := cfg.GetManager()
		logger := cfg.GetLogger()
		task.Client.Manager = mgr
		task.Client.Logger = logger
		if err = mgr.Start(); err != nil {
			logx.Errorf("start error: %v", err)
			abort()
			_ = task.Client.Manager.Stop()
			return
		}
		task.Client.HasStarted = true
		err = mgr.Wait()
		if err != nil {
			logx.Errorf("task wait error: %v", err)
			_ = GetTaskMgr().AbortTask(taskID, err.Error())
			_ = task.Client.Manager.Stop()
			return
		}
		if task.TaskInfo.TaskStatus == types.Running.String() {
			var status string
			if task.TaskInfo.Cron != "" {
				status = types.Scheduling.String()
			} else {
				status = types.Finished.String()
			}
			_ = GetTaskMgr().FinishTask(taskID, status)
		}
		_ = task.Client.Manager.Stop()
		signal <- struct{}{}
	}()
	return nil
}

func DeleteImportTask(tasksDir, taskID, address, username string) error {
	_, err := taskmgr.db.FindTaskInfoByIdAndAddressAndUser(taskID, address, username)
	if err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	err = GetTaskMgr().DelTask(tasksDir, taskID)
	if err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return nil
}

func GetImportTask(taskID, address, username string) (*types.GetImportTaskData, error) {
	_, err := taskmgr.db.FindTaskInfoByIdAndAddressAndUser(taskID, address, username)
	if err != nil {
		return nil, errors.New("task not existed")
	}
	if t, ok := GetTaskMgr().GetTask(taskID); ok {
		return t.ToImportTaskData()
	}
	return nil, errors.New("task not existed")
}

func GetManyImportTask(
	address, username, space string,
	pageIndex, pageSize int,
	status string) (*types.GetManyImportTaskData, error) {
	result := &types.GetManyImportTaskData{
		Total: 0,
		List:  []types.GetImportTaskData{},
	}

	tasks, count, err := taskmgr.db.FindTaskInfoByAddressAndUser(address, username, space, status, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		importAddress, err := parseImportAddress(t.ImportAddress)
		if err != nil {
			return nil, err
		}
		stats := t.Stats
		data := types.GetImportTaskData{
			Id:            t.BID,
			Status:        t.TaskStatus,
			Message:       t.TaskMessage,
			CreateTime:    t.CreateTime.UnixMilli(),
			UpdateTime:    t.UpdateTime.UnixMilli(),
			Address:       t.Address,
			ImportAddress: importAddress,
			User:          t.User,
			Name:          t.Name,
			Space:         t.Space,
			RawConfig:     t.RawConfig,
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
		result.List = append(result.List, data)
	}
	result.Total = count

	return result, nil
}

func StopImportTask(taskID, address, username string) error {
	_, err := taskmgr.db.FindTaskInfoByIdAndAddressAndUser(taskID, address, username)
	if err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}

	err = GetTaskMgr().StopTask(taskID)
	if err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	} else {
		return nil
	}
}

func parseImportAddress(address string) ([]string, error) {
	re := regexp.MustCompile(`,\s*`)
	split := re.Split(address, -1)
	importAddress := append([]string{}, split...)

	return importAddress, nil
}
