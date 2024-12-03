package importer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dcron-contrib/commons/dlog"
	"github.com/dcron-contrib/redisdriver"
	"github.com/libi/dcron"
	configv3 "github.com/lucky-xin/nebula-importer/pkg/config/v3"
	"github.com/lucky-xin/nebula-importer/pkg/task/db"
	"github.com/lucky-xin/nebula-importer/pkg/task/ecode"
	"github.com/lucky-xin/nebula-importer/pkg/task/types"
	"github.com/lucky-xin/nebula-importer/pkg/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	taskmgr *TaskMgr
)

func NewTaskMgr(serverName string, tackConfig *types.TaskConfig) {
	logger := &dlog.StdLogger{
		Log:        log.New(os.Stdout, "["+serverName+"]", log.LstdFlags),
		LogVerbose: false,
	}
	zone := time.FixedZone("UTC", 8*60*60)
	dcronInstance := dcron.NewDcronWithOption(
		serverName,
		redisdriver.NewDriver(tackConfig.CreateRedis()),
		dcron.WithLogger(logger),
		dcron.CronOptionLocation(zone),
		dcron.CronOptionSeconds(),
		dcron.WithHashReplicas(10),
		dcron.WithNodeUpdateDuration(time.Second*10),
	)
	openDB, err := tackConfig.CreateDb()
	if err != nil {
		logx.Errorf("open db error: %v", err)
		return
	}
	taskmgr = &TaskMgr{
		cache: &sync.Map{},
		db:    &db.TaskDb{DB: openDB},
		dcron: dcronInstance,
	}
	InitTask()
	go taskmgr.startCronTask()
}

type TaskMgr struct {
	cache  *sync.Map
	db     *db.TaskDb
	config *types.TaskConfig
	dcron  *dcron.Dcron
}

func InitTask() {
	if err := GetTaskMgr().db.UpdateProcessingTasks2Aborted(); err != nil {
		logx.Errorf("update processing cache to aborted failed: %s", err)
		panic(err)
	}
}

func CreateNewTaskDir(rootDir string, id string) (string, error) {
	taskDir := filepath.Join(rootDir, id)
	if err := utils.CreateDir(taskDir); err != nil {
		return "", ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return taskDir, nil
}

func CreateConfigFile(taskDir string, cfgBytes []byte) (string, error) {
	fileName := "config.yaml"
	path := filepath.Join(taskDir, fileName)
	confv3 := &configv3.Config{}
	err := json.Unmarshal(cfgBytes, confv3)
	if err != nil {
		return "", ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	// erase user information
	cfg := confv3
	cfg.Client.User = "${YOUR_NEBULA_NAME}"
	cfg.Client.Password = "${YOUR_NEBULA_PASSWORD}"
	cfg.Client.Address = "${YOUR_NEBULA_ADDRESS}"
	for _, source := range cfg.Sources {
		S3Config := source.S3
		SFTPConfig := source.SFTP
		OSSConfig := source.OSS
		if S3Config != nil {
			S3Config.AccessKeyID = "${YOUR_S3_ACCESS_KEY}"
			S3Config.AccessKeySecret = "${YOUR_S3_SECRET_KEY}"
		}
		if SFTPConfig != nil {
			SFTPConfig.User = "${YOUR_SFTP_USER}"
			SFTPConfig.Password = "${YOUR_SFTP_PASSWORD}"
		}
		if OSSConfig != nil {
			OSSConfig.AccessKeyID = "${YOUR_OSS_ACCESS_KEY}"
			OSSConfig.AccessKeySecret = "${YOUR_OSS_SECRET_KEY}"
		}
	}
	outYaml, err := yaml.Marshal(confv3)
	if err != nil {
		return "", ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	if err := os.WriteFile(path, outYaml, 0o644); err != nil {
		return "", ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return string(outYaml), nil
}

func (mgr *TaskMgr) NewTask(id, cron, host, user, taskName string, rawConfig string, cfg *types.ImportTaskV2Config) (*Task, error) {
	if t, err := mgr.GetTaskInfoByName(taskName); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if t != nil {
		return nil, ecode.WithErrorMessage(ecode.ErrForbidden, errors.New("task exists"))
	}
	// init task db
	taskInfo := &db.TaskInfo{
		BID:           id,
		Cron:          cron,
		Name:          taskName,
		Address:       host,
		Space:         cfg.Manager.GraphName,
		TaskStatus:    types.Running.String(),
		ImportAddress: cfg.Client.Address,
		User:          user,
		RawConfig:     rawConfig,
	}
	if cron != "" {
		taskInfo.TaskStatus = types.Scheduling.String()
	}

	if err := mgr.db.InsertTaskInfo(taskInfo); err != nil {
		return nil, err
	}
	task := &Task{
		Client: &Client{
			Cfg:        cfg,
			HasStarted: false,
		},
		TaskInfo: taskInfo,
	}
	mgr.PutTask(id, task)
	return task, nil
}

func (mgr *TaskMgr) StartTask(info *db.TaskInfo) error {
	if info.Cron != "" {
		mgr.dcron.Remove(info.Name)
		// start Cron import task
		err := mgr.dcron.AddFunc(info.Name, info.Cron, func() {
			if err := StartImport(info.BID); err != nil {
				logx.Error("exec Cron import task error, id: %s, Cron: %s", info.ID, err.Error())
			}
		})
		if err != nil {
			logx.Error("add Cron import task error, id: %s, Cron: %s", info.ID, err.Error())
			_ = GetTaskMgr().AbortTask(info.BID, err.Error())
			return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
		}
		logx.Info("start Cron import task, id: %s, Cron: %s", info.BID, info.Cron)
	} else {
		// start import
		if err := StartImport(info.BID); err != nil {
			logx.Errorf("add import task error, id: %d, Cron: %s", info.ID, err.Error())
			_ = GetTaskMgr().AbortTask(info.BID, err.Error())
			return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
		}
	}
	return nil
}

/*
StopTask will change the task status to `Stopped`,
and then call FinishTask
*/
func (mgr *TaskMgr) StopTask(taskID string) error {
	if task, ok := mgr.GetTask(taskID); ok {
		var err error
		manager := task.Client.Manager
		if manager != nil {
			if task.Client.HasStarted {
				err = manager.Stop()
				mgr.dcron.Remove(task.TaskInfo.Name)
			} else {
				// hack import not support stop before start()
				err = errors.New("task has not started, please try later")
			}
		} else {
			err = errors.New("manager is nil, please try later")
		}

		if err != nil {
			return fmt.Errorf("stop task failed: %w", err)
		}
		if err = mgr.FinishTask(taskID, types.Stopped.String()); err != nil {
			return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
		}
		return nil
	}
	return errors.New("task is finished or not exist")
}

/*
FinishTask will query task stats
  - delete task in the map
  - update taskInfo in db
  - update taskEffect in db
*/
func (mgr *TaskMgr) FinishTask(taskID, status string) (err error) {
	task, ok := mgr.GetTask(taskID)
	if !ok {
		return
	}
	defer mgr.cache.Delete(taskID)
	if err = task.UpdateQueryStats(); err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	task.TaskInfo.TaskStatus = status
	err = mgr.db.UpdateTaskInfo(task.TaskInfo)
	if err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return mgr.StorePartTaskLog(taskID)
}

func (mgr *TaskMgr) AbortTask(taskID, msg string) (err error) {
	task, ok := mgr.GetTask(taskID)
	if !ok {
		return nil
	}
	defer mgr.cache.Delete(taskID)
	task.TaskInfo.TaskStatus = types.Aborted.String()
	task.TaskInfo.TaskMessage = msg
	err = mgr.db.UpdateTaskInfo(task.TaskInfo)
	if err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return mgr.StorePartTaskLog(taskID)
}

func (mgr *TaskMgr) GetTaskInfoByName(taskName string) (task *db.TaskInfo, err error) {
	return mgr.db.FindTaskInfoByName(taskName)
}

func (mgr *TaskMgr) TurnDraftToTask(id, cron, taskName string, rawCfg string, cfg *types.ImportTaskV2Config) (*Task, error) {
	// init task db
	taskInfo := &db.TaskInfo{
		BID:           id,
		Cron:          cron,
		Name:          taskName,
		Space:         cfg.Manager.GraphName,
		TaskStatus:    types.Running.String(),
		ImportAddress: cfg.Client.Address,
		RawConfig:     rawCfg,
		CreateTime:    time.Now(),
	}
	if cron != "" {
		taskInfo.TaskStatus = types.Scheduling.String()
	}
	if err := mgr.db.UpdateTaskInfo(taskInfo); err != nil {
		return nil, err
	}

	task := &Task{
		Client: &Client{
			Cfg:        cfg,
			Manager:    nil,
			Logger:     nil,
			HasStarted: false,
		},
		TaskInfo: taskInfo,
	}

	mgr.PutTask(id, task)
	return task, nil
}

func (mgr *TaskMgr) NewTaskDraft(id, host, user, taskName, space, rawCfg string) error {
	// init task db
	taskInfo := &db.TaskInfo{
		BID:        id,
		Name:       taskName,
		Address:    host,
		Space:      space,
		TaskStatus: types.Draft.String(),
		User:       user,
		RawConfig:  rawCfg,
	}
	if err := mgr.db.InsertTaskInfo(taskInfo); err != nil {
		return err
	}
	return nil
}

func (mgr *TaskMgr) NewTaskEffect(taskEffect *db.TaskEffect) error {
	if err := mgr.db.InsertTaskEffect(taskEffect); err != nil {
		return err
	}
	return nil
}

func (mgr *TaskMgr) UpdateTaskDraft(id, taskName, space, rawCfg string) error {
	// init task db
	taskInfo := &db.TaskInfo{
		BID:        id,
		Name:       taskName,
		Space:      space,
		TaskStatus: types.Draft.String(),
		RawConfig:  rawCfg,
	}
	if err := mgr.db.UpdateTaskInfo(taskInfo); err != nil {
		return err
	}
	return nil
}

func GetTaskMgr() *TaskMgr {
	return taskmgr
}

/*
GetTask get task from map and local sql
*/
func (mgr *TaskMgr) GetTask(taskID string) (*Task, bool) {
	if task, ok := mgr.getTaskFromMap(taskID); ok {
		return task, true
	}
	task := mgr.getTaskFromSQL(taskID)
	// did not find task
	if task.TaskInfo.ID == 0 {
		return nil, false
	}
	if task.Client == nil {
		c := types.ImportTaskV2Config{}
		byts, err := utils.Decrypt(task.TaskInfo.RawConfig, []byte(types.Cipher))
		if err != nil {
			logx.Errorf("Decrypt raw config error: %v", err)
			return nil, false
		}
		err = json.Unmarshal(byts, &c)
		if err != nil {
			logx.Errorf("Unmarshal raw config error: %v", err)
			return nil, false
		}
		task.Client = &Client{
			Cfg:        &c,
			HasStarted: false,
		}
	}
	mgr.PutTask(taskID, task)
	return task, true
}

/*
PutTask put task into cache map
*/
func (mgr *TaskMgr) PutTask(taskID string, task *Task) {
	mgr.cache.Store(taskID, task)
}

func (mgr *TaskMgr) DelTask(tasksDir, taskID string) error {
	_, ok := mgr.GetTask(taskID)
	if ok {
		go mgr.cache.Delete(taskID)
	}
	if err := mgr.db.DelTaskInfo(taskID); err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	taskDir := filepath.Join(tasksDir, taskID)
	return os.RemoveAll(taskDir)
}

func (mgr *TaskMgr) StorePartTaskLog(taskID string) error {
	filePath := filepath.Join(mgr.config.Dir.TasksDir, taskID, "import.log")
	content, err := utils.ReadPartFile(filePath)
	if err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	taskEffect := &db.TaskEffect{
		BID:        taskID,
		Log:        strings.Join(content, "\n"),
		CreateTime: time.Now(),
	}
	if err := mgr.db.UpdateTaskEffect(taskEffect); err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return nil
}

/*
UpdateTaskInfo will query task stats, update task in the map
and update the taskInfo in local sql
*/
func (mgr *TaskMgr) UpdateTaskInfo(taskID string) error {
	task, ok := mgr.GetTask(taskID)
	if !ok {
		return nil
	}
	return mgr.UpdateTask(task)
}

func (mgr *TaskMgr) UpdateTask(task *Task) error {
	if err := task.UpdateQueryStats(); err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return mgr.db.UpdateTaskInfo(task.TaskInfo)
}

// UpdateTaskInfoById Update task info by task id
func (mgr *TaskMgr) UpdateTaskInfoById(task *Task) error {
	return mgr.db.UpdateTaskInfo(task.TaskInfo)
}

func (mgr *TaskMgr) Stop() {
	mgr.dcron.Stop()
}

func (mgr *TaskMgr) getTaskFromMap(taskID string) (*Task, bool) {
	if t, ok := mgr.cache.Load(taskID); ok {
		return t.(*Task), true
	}
	return nil, false
}

func (mgr *TaskMgr) getTaskFromSQL(taskID string) *Task {
	taskInfo, err := mgr.db.FindTaskInfoById(taskID)
	if err != nil {
		logx.Error(err)
		return nil
	}
	task := new(Task)
	task.TaskInfo = taskInfo
	return task
}

func (mgr *TaskMgr) GetDcron() *dcron.Dcron {
	return mgr.dcron
}

func (mgr *TaskMgr) startCronTask() {
	mgr.dcron.Start()
	pageIndex := 1
	pageSize := 50
	for {
		tasks, err := mgr.db.PageCronTask(pageIndex, pageSize)
		if err != nil {
			logx.Error("page Cron task error:", err.Error())
			break
		}
		if len(tasks) == 0 {
			break
		}
		pageIndex++
		for _, task := range tasks {
			if task.TaskStatus == types.Scheduling.String() {
				go func() {
					_ = mgr.StartTask(task)
				}()
			}
		}
	}
}
