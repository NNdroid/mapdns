package db_entity

import "gorm.io/gorm"

type DnsRecord struct {
	gorm.Model
	ID          uint64 `json:"ID" gorm:"primaryKey;autoIncrement:true;not null"`
	RequestType uint16 `json:"RequestType" gorm:"uniqueIndex:idx_rtd;not null"`
	Domain      string `json:"Domain" gorm:"uniqueIndex:idx_rtd;not null"`
	Address     string `json:"Address"`
	Enable      bool   `json:"Enable" gorm:"default:true;not null"`
}
