package models

import (
	"log"
	"strconv"

	"github.com/ducksouplab/mastok/otree"
	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	CampaignId uint
	// otree Code
	Code           string
	OtreeId        string
	OtreeCreatedAt string
	Config         string
	Size           int
	AdminUrl       string
}

func convertFromOtree(o otree.Session) Session {
	return Session{
		Code:           o.Code,
		OtreeId:        o.Config.Id,
		OtreeCreatedAt: o.FormatCreatedAt(),
		Config:         o.ConfigName,
		Size:           o.NumParticipants,
		AdminUrl:       o.AdminUrl,
	}
}

func NewSession(c *Campaign) (session Session, participantCodes []string, err error) {
	sessionId := "mk:" + c.Namespace + ":" + strconv.Itoa(c.StartedSessions+1)
	args := otree.SessionArgs{
		ConfigName:      c.Config,
		NumParticipants: c.PerSession,
		Config: otree.NestedConfig{
			Id: sessionId,
		},
	}
	o := otree.Session{}
	// GET code
	if err = otree.PostOTreeJSON("/api/sessions/", args, &o); err != nil {
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
	err = appendSessionToCampaign(c, session)
	if err != nil {
		log.Println("[runner] add session to campaign failed: ", err)
	}
	return
}
