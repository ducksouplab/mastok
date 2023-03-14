package models

import (
	"github.com/ducksouplab/mastok/env"
	"gorm.io/gorm"
)

const (
	Paused    string = "Paused"
	Running   string = "Running"
	Completed string = "Completed"
)

// Caution: validations are done when binding (in handlers), before and not related to gorm
type Campaign struct {
	gorm.Model
	Namespace          string `form:"namespace" binding:"required,alphanum,min=2,max=128" gorm:"uniqueIndex"`
	Slug               string `form:"slug" binding:"required,alphanum,min=2,max=128" gorm:"uniqueIndex"`
	Info               string `form:"info" binding:"max=128"`
	Config             string `form:"config" binding:"required"`
	State              string `gorm:"default:Paused"`
	PerSession         int    `form:"per_session" binding:"required,gte=1,lte=32"`
	MaxSessions        int    `form:"max_sessions" binding:"required,gte=1,lte=32"`
	ConcurrentSessions int    `form:"concurrent_sessions" binding:"required,gte=1,lte=32" gorm:"default:1"`
	StartedSessions    int    `gorm:"default:0"`
	// relations
	Sessions []Session
}

func FindCampaignByNamespace(namespace string) (c *Campaign, err error) {
	err = DB.First(&c, "namespace = ?", namespace).Error
	return
}

func FindCampaignByNamespaceWithSessions(namespace string) (c *Campaign, err error) {
	err = DB.Preload("Sessions").First(&c, "namespace = ?", namespace).Error
	return
}

func appendSessionToCampaign(c *Campaign, s Session) (err error) {
	c.StartedSessions += 1
	if c.StartedSessions == c.MaxSessions {
		c.State = "Completed"
	}
	c.Sessions = append(c.Sessions, s)
	err = DB.Save(c).Error
	return
}

func (c *Campaign) FormatCreatedAt() string {
	return c.CreatedAt.Format("2006-01-02 15:04:05")
}

func (c *Campaign) ShareURL() string {
	return env.Origin + "/join/" + c.Slug
}
