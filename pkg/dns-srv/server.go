package dns_srv

import (
	"context"
	"github.com/miekg/dns"
	"mapdns/pkg/cache-srv"
	"mapdns/pkg/config"
	"mapdns/pkg/db-srv"
	"mapdns/pkg/log"
)

type Server struct {
	cfg   *config.Config
	ctx   context.Context
	srv   *dns.Server
	cache *cache_srv.Server
	db    *db_srv.Server
}

func New(ctx context.Context, cfg *config.Config, db *db_srv.Server, cache *cache_srv.Server) *Server {
	srv := &Server{cfg: cfg, ctx: ctx, db: db, cache: cache}
	err := srv.cache.ClearRecords()
	if err != nil {
		log.Logger().Fatal(err)
	}
	srv.loadToRedis()
	return srv
}

func (srv *Server) loadToRedis() {
	dat, err := srv.db.GetAvailableDNSRecords()
	if err != nil {
		log.Logger().Errorf("Error load available dns records: %v", err)
	}
	for _, record := range dat {
		err := srv.cache.SetRecord(record.RequestType, record.Domain, record.Address)
		if err != nil {
			log.Logger().Fatalf("failed to load record: %v", err)
		}
	}
}

func (srv *Server) ListenAndServe() error {
	srv.srv = &dns.Server{Addr: srv.cfg.DNS.Listen, Net: "udp"}
	srv.srv.Handler = &Handler{srv: srv}
	if err := srv.srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
