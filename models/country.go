package models

type Country struct {
	Generic

	Name string `gorm:"not null; unique" json:"name"`
	Code string `gorm:"not null; unique" json:"code"`

	ServersAvailable int `gorm:"<-:false;->;-:migration" json:"servers_available"`
}
