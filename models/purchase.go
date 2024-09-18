package models

type Purchase struct {
	Generic

	EventId string `gorm:"not null; unique" json:"event_id"`

	Address string `gorm:"not null" json:"address"`
	Amount  int64  `gorm:"not null" json:"amount"`
	Denom   string `gorm:"not null" json:"denom"`

	IsRedeemed bool `gorm:"not null; default:false" json:"is_redeemed"`
}
