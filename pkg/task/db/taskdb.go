package db

import (
	"github.com/lucky-xin/nebula-importer/pkg/task/ecode"
	"github.com/lucky-xin/nebula-importer/pkg/task/types"
	"gorm.io/gorm"
)

type TaskDb struct {
	*gorm.DB
}

// FindTaskInfoByIdAndAddressAndUser used to check whether the task belongs to the user
func (t *TaskDb) FindTaskInfoByIdAndAddressAndUser(id, address, user string) (*TaskInfo, error) {
	taskInfo := new(TaskInfo)
	if err := t.Model(&TaskInfo{}).
		Where("b_id = ? AND address = ? And user = ?", id, address, user).First(&taskInfo).Error; err != nil {
		return nil, err
	}
	return taskInfo, nil
}

func (t *TaskDb) FindTaskInfoById(id string) (*TaskInfo, error) {
	taskInfo := new(TaskInfo)
	if err := t.Model(&TaskInfo{}).Where("b_id = ?", id).First(&taskInfo).Error; err != nil {
		return nil, err
	}
	return taskInfo, nil
}

func (t *TaskDb) FindTaskInfoByAddressAndUser(space, status string, pageIndex, pageSize int) ([]*TaskInfo, int64, error) {
	tasks := make([]*TaskInfo, 0)
	var count int64
	tx := t.Model(&TaskInfo{})
	if space != "" {
		tx = tx.Where("space = ?", space)
	}
	if status != "" {
		tx = tx.Where("task_status = ?", status)
	}
	tx = tx.Order("id desc")
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, count, err
	}
	return tasks, count, nil
}

func (t *TaskDb) FindTaskInfoByName(name string) (*TaskInfo, error) {
	taskInfo := new(TaskInfo)
	err := t.Model(&TaskInfo{}).Where("name = ?", name).First(&taskInfo).Error
	if err != nil {
		return nil, err
	}
	return taskInfo, nil
}

func (t *TaskDb) PageCronTask(pageIndex, pageSize int) (tasks []*TaskInfo, err error) {
	tx := t.Model(&TaskInfo{}).Where("cron != ''")
	tx = tx.Order("id desc")
	err = tx.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return
	}
	return
}

func (t *TaskDb) InsertTaskInfo(info *TaskInfo) error {
	return t.Create(info).Error
}

func (t *TaskDb) UpdateTaskInfo(info *TaskInfo) error {
	if info.TaskStatus == types.Finished.String() {
		stats := info.Stats
		if stats.Processed != stats.Total {
			stats.Processed = stats.Total
		}
	}
	return t.Model(&TaskInfo{}).Where("b_id = ?", info.BID).Updates(info).Error
}

func (t *TaskDb) DelTaskInfo(ID string) error {
	_ = t.Delete(&TaskEffect{}, "task_id = ?", ID).Error
	return t.Delete(&TaskInfo{}, "b_id = ?", ID).Error
}

func (t *TaskDb) UpdateProcessingTasks2Aborted() error {
	if err := t.Model(&TaskInfo{}).Where("task_status = ?", types.Running.String()).
		Updates(&TaskInfo{TaskStatus: types.Aborted.String(), TaskMessage: "Service exception"}).Error; err != nil {
		return ecode.WithErrorMessage(ecode.ErrInternalServer, err)
	}
	return nil
}

func (t *TaskDb) InsertTaskEffect(taskEffect *TaskEffect) error {
	return t.Create(taskEffect).Error
}

func (t *TaskDb) UpdateTaskEffect(taskEffect *TaskEffect) error {
	return t.Model(&TaskEffect{}).Where("task_id = ?", taskEffect.BID).Updates(taskEffect).Error
}

func (t *TaskDb) DelTaskEffect(ID string) error {
	return t.Delete(&TaskEffect{}, "task_id = ?", ID).Error
}
