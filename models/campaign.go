package models

import (
	"gorm.io/gorm"
)

type CampaignState int

const (
	Waiting = iota
	Running
	Paused
	Complete
)

func (s CampaignState) String() string {
	return [...]string{"Waiting", "Running", "Paused", "Complete"}[s]
}

// Caution: validations are done when binding (in handlers), before and not related to gorm
type Campaign struct {
	gorm.Model
	Namespace        string        `form:"namespace" binding:"required,alphanum,min=2,max=128" gorm:"uniqueIndex"`
	Slug             string        `form:"slug" binding:"required,alphanum,min=2,max=128" gorm:"uniqueIndex"`
	Info             string        `form:"info" binding:"max=128"`
	ExperimentConfig string        `form:"experiment_config" binding:"required"`
	PerSession       uint          `form:"per_session" binding:"required,gte=1,lte=32"`
	SessionsMax      uint          `form:"sessions_max" binding:"required,gte=1,lte=32"`
	State            CampaignState `gorm:"default:0"`
	SessionsStarted  uint          `gorm:"default:0"`
}

func FindCampaignByNamespace(namespace string) (c *Campaign, err error) {
	err = DB.First(&c, "namespace = ?", namespace).Error
	return
}

func (c *Campaign) FormatCreatedAt() string {
	return c.CreatedAt.Format("2006-01-02 15:04:05")
}
