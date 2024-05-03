package db_srv

import (
	"github.com/glebarez/sqlite"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"mapdns/pkg/config"
	"mapdns/pkg/db-entity"
	"mapdns/pkg/log"
	"strings"
)

type Server struct {
	ctx context.Context
	cfg *config.Config
	db  *gorm.DB
}

func New(ctx context.Context, cfg *config.Config) *Server {
	srv := Server{ctx: ctx, cfg: cfg}
	srv.initDB()
	return &srv
}

func (srv *Server) initDB() {
	var err error
	srv.db, err = gorm.Open(sqlite.Open(srv.cfg.DB.Path), &gorm.Config{})
	if err != nil {
		log.Logger().Fatalf("failed to open database: %v", err)
	}
	err = srv.db.AutoMigrate(&db_entity.DnsRecord{})
	if err != nil {
		log.Logger().Fatalf("failed to migrate database: %v", err)
	}
}

func (srv *Server) closeDB() error {
	return nil
}

func (srv *Server) GetAvailableDNSRecords() ([]*db_entity.DnsRecord, error) {
	var records []*db_entity.DnsRecord
	srv.db.Find(&records, "enable = ?", 1)
	return records, nil
}

func (srv *Server) InsertRecord(records []*db_entity.DnsRecord) error {
	for _, record := range records {
		if !strings.HasSuffix(record.Domain, ".") {
			record.Domain = record.Domain + "."
		}
	}
	srv.db.Create(records)
	return nil
}

func (srv *Server) DeleteRecordById(ids []uint64) error {
	srv.db.Delete(&db_entity.DnsRecord{}, ids)
	return nil
}

func (srv *Server) GetRecordById(id uint64) (*db_entity.DnsRecord, error) {
	var record db_entity.DnsRecord
	srv.db.First(&record, id)
	return &record, nil
}

func (srv *Server) GetRecordListById(id []uint64) ([]*db_entity.DnsRecord, error) {
	var records []*db_entity.DnsRecord
	srv.db.Where("id IN (?)", id).Find(&records)
	return records, nil
}
