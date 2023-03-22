package models

import (
	"errors"
	"log"

	"gorm.io/gorm"
)

type Participation struct {
	gorm.Model
	CampaignID  uint   `gorm:"index"`
	SessionID   uint   `gorm:"index"`
	Fingerprint string `gorm:"index"`
	OtreeCode   string
}

func CreateParticipation(s Session, fingerprint, code string) (err error) {
	participation := Participation{
		CampaignID:  s.CampaignID,
		SessionID:   s.ID,
		Fingerprint: fingerprint,
		OtreeCode:   code,
	}
	return DB.Create(&participation).Error
}

// ok if participation exists, live if participation is related to a live session
func GetParticipation(c Campaign, fingerprint string) (p *Participation, ok bool) {
	err := DB.First(&p, "campaign_id = ? AND fingerprint = ?", c.ID, fingerprint).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("[db] error: ", err)
		}
		return
	}
	ok = true
	return
}
