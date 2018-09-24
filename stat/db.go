package stat

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type DB interface {
	Insert(metrics) error
}

type Mysql struct {
	db *gorm.DB
}

func NewMysql(username, passwd, uri, database string) (*Mysql, error) {
	args := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True", username, passwd, uri, database)
	db, err := gorm.Open("mysql", args)
	if err != nil {
		return nil, err
	}

	return &Mysql{
		db: db,
	}, nil
}

func (m *Mysql) Insert(met metrics) error {
	return m.db.Create(met).Error
}
