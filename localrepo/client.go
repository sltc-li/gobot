package localrepo

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type Repository interface {
	Migrate(data interface{}) error
	Put(data interface{}) error
	Del(cond interface{}) error
	GetOne(cond interface{}, data interface{}) error
	GetAll(cond interface{}, data interface{}) error
	Close() error
}

func New() (Repository, error) {
	db, err := gorm.Open("sqlite3", "./.bot.sqlite")
	if err != nil {
		return nil, err
	}
	return &repo{db: db}, nil
}

type repo struct {
	db *gorm.DB
}

func (r *repo) Migrate(data interface{}) error {
	return r.db.AutoMigrate(data).Error
}

func (r *repo) Put(data interface{}) error {
	return r.db.Create(data).Error
}

func (r *repo) Del(cond interface{}) error {
	return r.db.Where(cond).Delete(cond).Error
}

func (r *repo) GetOne(cond interface{}, data interface{}) error {
	return r.db.Where(cond).First(data).Error
}

func (r *repo) GetAll(cond interface{}, data interface{}) error {
	return r.db.Where(cond).Find(data).Error
}

func (r *repo) Close() error {
	return r.db.Close()
}
