package model

type CcsApplication struct {
	AppId      string `gorm:"primaryKey"`
	Name       string
	MQEndpoint string
	SecretKey  string
}

type CcsService struct {
	AppId     string `gorm:"primaryKey"`
	ServiceId string `gorm:"primaryKey"`
	Name      string
}

type CcsGrid struct {
	ServiceId       string `gorm:"primaryKey"` // ServiceId
	ParentServiceId string `gorm:"primaryKey"` // Parent
}
