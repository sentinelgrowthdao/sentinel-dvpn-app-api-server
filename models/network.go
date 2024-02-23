package models

type Network struct {
	Network   string `gorm:"type:cidr; unique"`
	Latitude  float64
	Longitude float64
}
