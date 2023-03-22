package models

import (
	"errors"

	"github.com/ducksouplab/mastok/env"
	"gorm.io/gorm"
)

// Paused/Running/Completed are the only values meant to be persisted.
// Regarding the two other derived states:
// - "Busy" is shared with supervisors when campaign is "Running" but ConcurrentSessions limit is also reached (by currentSessions())
// - "Unavailable" is shared to participants when state is not "Running", since they don't need the details
const (
	Paused      string = "Paused"
	Running     string = "Running"
	Completed   string = "Completed"
	Busy        string = "Busy"
	Unavailable string = "Unavailable"
)

// Caution: validations are done when binding (in routes)
// with https://github.com/go-playground/validator
// it's not related to gorm
type Campaign struct {
	gorm.Model
	Namespace          string `form:"namespace" binding:"required,namespace,min=2,max=128" gorm:"uniqueIndex"`
	Slug               string `form:"slug" binding:"required,namespace,min=2,max=128" gorm:"uniqueIndex"`
	Info               string `form:"info" binding:"max=128"`
	OtreeExperiment    string `form:"otree_experiment_id" binding:"required"`
	PerSession         int    `form:"per_session" binding:"required,gte=1,lte=32"`
	JoinOnce           bool   `form:"join_once" gorm:"default:false" ` // don't require due to https://github.com/go-playground/validator/issues/1040
	MaxSessions        int    `form:"max_sessions" binding:"required,gte=1,lte=32"`
	ConcurrentSessions int    `form:"concurrent_sessions" binding:"required,gte=1,lte=99" gorm:"default:1"`
	SessionDuration    int    `form:"session_duration" binding:"required"`
	State              string `gorm:"default:Paused"`
	StartedSessions    int    `gorm:"default:0"`
	// relations
	Sessions []Session
}

func FindCampaignByNamespace(namespace string) (c *Campaign, err error) {
	err = DB.Preload("Sessions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sessions.created_at DESC")
	}).First(&c, "namespace = ?", namespace).Error
	return
}

func FindCampaignBySlug(slug string) (c *Campaign, err error) {
	err = DB.Preload("Sessions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sessions.created_at DESC")
	}).First(&c, "slug = ?", slug).Error
	return
}

func (c *Campaign) appendSession(s *Session) (err error) {
	if c.State == Completed {
		return errors.New("session can't be added to completed campaign")
	}
	// process session
	s.Duration = c.SessionDuration
	s.CampaignID = c.ID
	// do append
	c.StartedSessions += 1
	if c.StartedSessions == c.MaxSessions {
		c.State = "Completed"
	}
	// append to in memory struct
	c.Sessions = append(c.Sessions, *s)
	// updates campaign fields + save session
	err = DB.Save(c).Error
	// replace s to get gorm ID
	*s = c.Sessions[len(c.Sessions)-1]
	return
}

func (c *Campaign) isBusy() bool {
	return c.State == Running && (c.liveSessions() >= c.ConcurrentSessions)
}

func (c *Campaign) liveSessions() (count int) {
	for _, s := range c.Sessions {
		if s.IsLive() {
			count++
		}
	}
	return
}

func (c *Campaign) GetPublicState(isSupervisor bool) (state string) {
	// initializes
	if c.isBusy() {
		state = Busy
	} else {
		state = c.State
	}
	// filters if not supervisor
	if !isSupervisor && (state != Running) {
		return Unavailable
	}
	return
}

func (c *Campaign) FormatCreatedAt() string {
	return c.CreatedAt.Format("2006-01-02 15:04:05")
}

func (c *Campaign) ShareURL() string {
	return env.Origin + "/join/" + c.Slug
}
