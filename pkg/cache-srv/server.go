package cache_srv

import (
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"mapdns/pkg/config"
)

type Server struct {
	cfg *config.Config
	ctx context.Context
	rdb *redis.Client
}

func New(ctx context.Context, cfg *config.Config) *Server {
	srv := &Server{ctx: ctx, cfg: cfg}
	srv.initRedis()
	return srv
}
