package models

import (
	"bufio"
	"errors"
	"log"
	"strconv"
	"strings"

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
	// definition
	OtreeExperiment    string `form:"otree_experiment_id" binding:"required"`
	Namespace          string `form:"namespace" binding:"required,namespaceValidate,min=2,max=128" gorm:"uniqueIndex"`
	Slug               string `form:"slug" binding:"required,namespaceValidate,min=2,max=128" gorm:"uniqueIndex"`
	PerSession         int    `form:"per_session" binding:"required,gte=1,lte=32"`
	JoinOnce           bool   `form:"join_once" gorm:"default:false" ` // don't <require> due to https://github.com/go-playground/validator/issues/1040
	MaxSessions        int    `form:"max_sessions" binding:"required,gte=1,lte=32"`
	ConcurrentSessions int    `form:"concurrent_sessions" binding:"gte=1,lte=99" gorm:"default:1"`
	SessionDuration    int    `form:"session_duration" binding:"required" gorm:"default:10"`
	WaitingLimit       int    `form:"waiting_limit" binding:"gte=1,lte=12" gorm:"default:5"`
	// extra configuration
	Grouping string `form:"grouping" binding:"groupingValidate=PerSession" gorm:"size:1024"`
	Consent  string `form:"consent" gorm:"size:65535"`
	// evolving
	State           string `gorm:"default:Paused"`
	StartedSessions int    `gorm:"default:0"`
	// relations
	Sessions []Session
}

type Group struct {
	Label string
	Size  int
}

type Grouping struct {
	Question string
	Groups   []Group
}

func GetCampaignByNamespace(namespace string) (c *Campaign, ok bool) {
	err := DB.Preload("Sessions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sessions.created_at DESC")
	}).First(&c, "namespace = ?", namespace).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("[db] error: ", err)
		}
		return
	}
	ok = true
	return
}

func GetCampaignBySlug(slug string) (c *Campaign, ok bool) {
	err := DB.Preload("Sessions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sessions.created_at DESC")
	}).First(&c, "slug = ?", slug).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("[db] error: ", err)
		}
		return
	}
	ok = true
	return
}

// Does not compare groupingStr and PerSession (it's done in router/validators)
func RawParseGroupingString(groupingStr string) (grouping Grouping, err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
	scanner := bufio.NewScanner(strings.NewReader(groupingStr))
	var question string
	var groups []Group
	for scanner.Scan() {
		line := scanner.Text()
		if len(question) == 0 {
			question = line
		} else {
			splits := strings.Split(line, ":")
			size, err := strconv.Atoi(splits[1])
			if err != nil {
				return grouping, err
			}
			groups = append(groups, Group{splits[0], size})
		}
	}
	return Grouping{question, groups}, nil
}

// returns nil if empty or invalid
func (c *Campaign) GetGrouping() (grouping *Grouping) {
	if len(c.Grouping) == 0 {
		return nil
	} else {
		grouping, err := RawParseGroupingString(c.Grouping)
		if err != nil {
			return nil
		}
		return &grouping
	}
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

func (c *Campaign) IsBusy() bool {
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

// Busy is a temporary state, participants can wait
func (c *Campaign) IsLive() bool {
	return c.State == Running || c.IsBusy()
}

func (c *Campaign) GetLiveState() (state string) {
	if c.IsBusy() {
		return Busy
	}
	return c.State
}

func (c *Campaign) FormatCreatedAt() string {
	return c.CreatedAt.Format("2006-01-02 15:04:05")
}

func (c *Campaign) ShareURL() string {
	return env.Origin + "/join/" + c.Slug
}
