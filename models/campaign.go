package models

import (
	"gorm.io/gorm"
)

const MASTOK_PREFIX = "mk:"

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
// Namespace min length is 2 + length(MASTOK_PREFIX)
type Campaign struct {
	gorm.Model
	Namespace        string        `form:"namespace" binding:"required,alphanum,min=2,max=64" gorm:"uniqueIndex"`
	Title            string        `form:"title" binding:"max=128"`
	ExperimentConfig string        `form:"experiment_config" binding:"required"`
	PerSession       uint          `form:"per_session" binding:"required,gte=1,lte=32"`
	SessionsMax      uint          `form:"sessions_max" binding:"required,gte=1,lte=32"`
	State            CampaignState `gorm:"default:0"`
	SessionsStarted  uint          `gorm:"default:0"`
}

func (c *Campaign) BeforeCreate(tx *gorm.DB) (err error) {
	c.Namespace = MASTOK_PREFIX + c.Namespace
	return
}

func (c *Campaign) FormatCreatedAt() string {
	return c.CreatedAt.Format("2006-01-02 15:04:05")
}
