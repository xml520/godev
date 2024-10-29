package godev

import (
	"fmt"
	"testing"
	"time"
)

type CertModel struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	Default         bool      `json:"default" gorm:"default:0;index"`
	Name            string    `json:"name" gorm:"comment:证书名称"`
	PublicName      string    `json:"public_name" gorm:"comment:外部名称"`
	CertPath        string    `json:"-"`
	Mobileprovision string    `json:"-"`
	Remark          string    `json:"remark" gorm:"comment:备注"`
	Status          int       `json:"status" gorm:"comment:状态;index"`
	CheckTime       time.Time `json:"check_time" gorm:"comment:最后检测"`
	IssueTime       time.Time `json:"issue_time"`
	ExpireTime      time.Time `json:"expire_time"`
	CreateTime      time.Time `json:"create_time" gorm:"comment:创建时间;autoCreateTime"`
	DeleteTime      time.Time `json:"-" gorm:"index"`
}

func TestName(t *testing.T) {
	generate, err := NewGenerate(new(CertModel))
	if err != nil {
		panic(err)
	}
	fmt.Println(generate.Names().ToSnake())
	for _, field := range generate.Fields {
		fmt.Println("name", field.Name(), "json", field.JsonName(), "comment ", field.Label(), "是否string", field.IsString())
	}
}
func TestNewRender(t *testing.T) {
	generate, err := NewGenerate(new(CertModel))
	if err != nil {
		panic(err)
	}
	err = NewRender("./*.gotpl", generate, false)
	if err != nil {
		panic(err)
	}
}
