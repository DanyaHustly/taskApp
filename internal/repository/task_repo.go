package repository

import (
	"errors"

	"gorm.io/gorm"
	"taskApp/internal/model"
)

type TaskRepository interface {
	Create(task *model.Task) error
	List() ([]model.Task, error)
	GetByID(id string) (*model.Task, error)
	Update(task *model.Task) error
	Delete(id string) error
}

type taskRepo struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepo{db: db}
}

func (r *taskRepo) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepo) List() ([]model.Task, error) {
	var tasks []model.Task
	if err := r.db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepo) GetByID(id string) (*model.Task, error) {
	var t model.Task
	if err := r.db.First(&t, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *taskRepo) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

func (r *taskRepo) Delete(id string) error {
	return r.db.Delete(&model.Task{}, "id = ?", id).Error
}
