package models

import (
	"gorm.io/gorm"
)

const MASTOK_PREFIX = "mk#"

// Caution: validations are done when binding (in handlers), before and not related to gorm
// Namespace min length is 2 + length(MASTOK_PREFIX)
type Campaign struct {
	gorm.Model
	Namespace        string `form:"namespace" binding:"required,alphanum,min=5,max=64" gorm:"uniqueIndex"`
	Title            string `form:"title" binding:"max=128"`
	ExperimentConfig string `form:"experiment_config" binding:"required"`
	PerSession       uint   `form:"per_session" binding:"required,gte=1,lte=32"`
	SessionMax       uint   `form:"session_max" binding:"required,gte=1,lte=32"`
}

func (c *Campaign) BeforeCreate(tx *gorm.DB) (err error) {
	c.Namespace = MASTOK_PREFIX + c.Namespace
	return
}
