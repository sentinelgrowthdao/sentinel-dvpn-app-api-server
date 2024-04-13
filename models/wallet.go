package models

type Wallet struct {
	Generic

	Address      string `gorm:"not null; unique" json:"address"`
	IsFeeGranted bool   `gorm:"index; not null; default:false" json:"is_fee_granted"`
}
