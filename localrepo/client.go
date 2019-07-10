package localrepo

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type Repository interface {
	Migrate(values ...interface{}) error
	Put(value interface{}) error
	Del(cond interface{}) error
	GetOne(where interface{}, out interface{}) error
	GetAll(where interface{}, out interface{}) error
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

func (r *repo) Migrate(values ...interface{}) error {
	return r.db.AutoMigrate(values...).Error
}

func (r *repo) Put(value interface{}) error {
	return r.db.Create(value).Error
}

func (r *repo) Del(where interface{}) error {
	return r.db.Where(where).Delete(where).Error
}

func (r *repo) GetOne(where interface{}, out interface{}) error {
	return r.db.Where(where).First(out).Error
}

func (r *repo) GetAll(where interface{}, out interface{}) error {
	return r.db.Where(where).Find(out).Error
}

func (r *repo) Close() error {
	return r.db.Close()
}
