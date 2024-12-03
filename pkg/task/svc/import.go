package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lucky-xin/nebula-importer/pkg/config"
	configv3 "github.com/lucky-xin/nebula-importer/pkg/config/v3"
	importer "github.com/lucky-xin/nebula-importer/pkg/task"
	"github.com/lucky-xin/nebula-importer/pkg/task/db"
	"github.com/lucky-xin/nebula-importer/pkg/task/ecode"
	"github.com/lucky-xin/nebula-importer/pkg/task/idx"
	"github.com/lucky-xin/nebula-importer/pkg/task/types"
	"github.com/lucky-xin/nebula-importer/pkg/utils"
	"github.com/zeromicro/go-zero/core/logx"
	"path"
	"path/filepath"
	"strings"
)

var (
	_ ImportService = (*importService)(nil)
)

const (
	importLogName = "import.log"
	errContentDir = "err"
)

type (
	ImportService interface {
		CreateImportTask(t *types.CreateImportTaskReq) (*types.CreateImportTaskData, error)
		RestartImportTask(t *types.RestartImportTaskReq) (bool, error)
		StopImportTask(request *types.StopImportTaskReq) error
		DeleteImportTask(*types.DeleteImportTaskReq) error
		GetImportTask(*types.GetImportTaskReq) (*types.GetImportTaskData, error)
		GetManyImportTask(request *types.GetManyImportTaskReq) (*types.GetManyImportTaskData, error)
	}

	importService struct {
		logx.Logger
		ctx              context.Context
		gormErrorWrapper utils.GormErrorWrapper
		idGenerator      idx.Generator
		config           *types.TaskConfig
	}
)

func NewImportService(ctx context.Context, config *types.TaskConfig) ImportService {
	return &importService{
		Logger:      logx.WithContext(ctx),
		ctx:         ctx,
		config:      config,
		idGenerator: idx.New(config.CreateRedis()),
	}
}

func updateConfig(confV3 *configv3.Config, taskDir, uploadDir string) {
	if confV3.Log == nil {
		confV3.Log = &config.Log{}
		confV3.Log.Files = make([]string, 0)
	}
	confV3.Log.Files = append(confV3.Log.Files, filepath.Join(taskDir, importLogName))
	dir := path.Dir(uploadDir)
	for _, s := range confV3.Sources {
		if s.Local != nil && !strings.HasPrefix(s.Local.Path, dir) {
			s.Local.Path = filepath.Join(dir, s.Local.Path)
		}
	}
}

func (i *importService) CreateImportTask(req *types.CreateImportTaskReq) (*types.CreateImportTaskData, error) {
	byts, err := json.Marshal(req.Config)
	if err != nil {
		return nil, ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return i.doCreateImportTask(req.Id, req.Name, byts, req.Config)
}

func (i *importService) doCreateImportTask(id *string, name string, cfgBytes []byte, conf *types.ImportTaskV2Config) (*types.CreateImportTaskData, error) {
	create := id == nil
	if create {
		newId := i.idGenerator.String()
		id = &newId
	}
	taskDir, err := importer.CreateNewTaskDir(i.config.Dir.TasksDir, *id)
	if err != nil {
		return nil, ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	// create config file
	configFile, err := importer.CreateConfigFile(taskDir, cfgBytes)
	if err != nil {
		return nil, ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	// modify source file path & add log config
	updateConfig(&conf.Config, taskDir, i.config.Dir.UploadDir)
	// init task in db
	host := i.config.Nebula.Address
	taskMgr := importer.GetTaskMgr()
	var task *importer.Task
	crypto, err := utils.Encrypt(cfgBytes, []byte(types.Cipher))
	if err != nil {
		return nil, fmt.Errorf("encrypt raw config error: %v", err)
	}
	if create {
		task, err = taskMgr.NewTask(*id, conf.Cron, host, i.config.Nebula.User, name, crypto, conf)
	} else {
		task, err = taskMgr.TurnDraftToTask(*id, conf.Cron, name, crypto, conf)
	}
	if err != nil {
		return nil, ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}

	// init task effect in db, store config.yaml
	err = taskMgr.NewTaskEffect(&db.TaskEffect{BID: *id, Config: configFile})
	if err != nil {
		return nil, ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	err = taskMgr.StartTask(task.TaskInfo)
	if err != nil {
		return nil, ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return &types.CreateImportTaskData{
		Id: task.TaskInfo.BID,
	}, nil
}

func (i *importService) RestartImportTask(t *types.RestartImportTaskReq) (bool, error) {
	// start import
	taskMgr := importer.GetTaskMgr()
	task, b := taskMgr.GetTask(*t.Id)
	if b {
		return false, errors.New("not found task,id:" + *t.Id)
	}
	err := taskMgr.StartTask(task.TaskInfo)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (i *importService) StopImportTask(req *types.StopImportTaskReq) error {
	host := i.config.Nebula.Address
	return importer.StopImportTask(req.Id, host, i.config.Nebula.User)
}

func (i *importService) DeleteImportTask(req *types.DeleteImportTaskReq) error {
	host := i.config.Nebula.Address
	return importer.DeleteImportTask(i.config.Dir.TasksDir, req.Id, host, i.config.Nebula.User)
}

func (i *importService) GetImportTask(req *types.GetImportTaskReq) (*types.GetImportTaskData, error) {
	return importer.GetImportTask(req.Id, i.config.Nebula.Address, i.config.Nebula.User)
}

func (i *importService) GetManyImportTask(req *types.GetManyImportTaskReq) (*types.GetManyImportTaskData, error) {
	host := i.config.Nebula.Address
	return importer.GetManyImportTask(host, i.config.Nebula.User, req.Space, req.Page, req.PageSize, req.Status)
}
