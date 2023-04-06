package models

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/ducksouplab/mastok/otree"
	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	CampaignID uint      `gorm:"index"`
	LaunchedAt time.Time // used to see if session is currently runnning
	Duration   int       // copied from Campaign SessionDuration
	// otree Code
	Code     string
	OtreeId  string
	Size     int
	AdminUrl string
}

const OtreePrefix = "mk:"

func convertFromOtree(o otree.Session) Session {
	return Session{
		Code:     o.Code,
		OtreeId:  o.Config.Id,
		Size:     o.NumParticipants,
		AdminUrl: o.AdminUrl,
	}
}

func newSessionArgs(c *Campaign) otree.SessionArgs {
	sessionId := OtreePrefix + c.Namespace + ":" + strconv.Itoa(c.StartedSessions+1)
	return otree.SessionArgs{
		SessionConfigName: c.OtreeExperiment,
		NumParticipants:   c.PerSession,
		Config: otree.NestedConfig{
			Id: sessionId,
		},
	}
}

func CreateSession(c *Campaign) (session Session, participantCodes []string, err error) {
	args := newSessionArgs(c)
	o := otree.Session{}

	// GET code
	if err = otree.PostOTreeJSON("/api/sessions", args, &o); err != nil {
		return
	}
	// GET more details (participant codes) and override s
	err = otree.GetOTreeJSON("/api/sessions/"+o.Code, &o)
	if err != nil {
		log.Println("[runner] get oTree sessions failed: ", err)
	}
	for _, p := range o.Participants {
		participantCodes = append(participantCodes, p.Code)
	}
	// save to campaign
	session = convertFromOtree(o)
	session.LaunchedAt = time.Now()
	err = c.appendSession(&session)
	if err != nil {
		log.Println("[runner] add session to campaign failed: ", err)
	}
	return
}

func (s *Session) FormatCreatedAt() string {
	return s.CreatedAt.Format("2006-01-02 15:04:05")
}

func GetSession(id uint) (s *Session, ok bool) {
	err := DB.First(&s, "id = ?", id).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("[db] error: ", err)
		}
		return
	}
	ok = true
	return
}

func (s *Session) IsLive() bool {
	span := time.Duration(s.Duration) * SessionDurationUnit
	limit := time.Now().Add(-span)
	return s.LaunchedAt.After(limit)
}
