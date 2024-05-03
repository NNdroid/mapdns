package http_srv

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"mapdns/pkg/cache-srv"
	"mapdns/pkg/config"
	"mapdns/pkg/db-entity"
	"mapdns/pkg/db-srv"
	"net/http"
)

type Server struct {
	cfg   *config.Config
	ctx   context.Context
	db    *db_srv.Server
	cache *cache_srv.Server
}

func New(ctx context.Context, cfg *config.Config, db *db_srv.Server, cache *cache_srv.Server) *Server {
	return &Server{ctx: ctx, cfg: cfg, db: db, cache: cache}
}

func (srv *Server) ListenAndServe() error {
	if !srv.cfg.Verbose {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	dnsGroup := r.Group("/dns")
	dnsGroup.GET("/records", func(c *gin.Context) {
		dat, err := srv.db.GetAvailableDNSRecords()
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
		}
		c.JSON(http.StatusOK, NewRespOK(dat))
	})
	dnsGroup.POST("/add", func(c *gin.Context) {
		var json struct {
			RequestType uint16 `json:"request_type" binding:"required"`
			Domain      string `json:"domain" binding:"required"`
			Address     string `json:"address" binding:"required"`
		}
		if err := c.ShouldBind(&json); err != nil {
			c.JSON(http.StatusBadRequest, NewRespFail(err.Error()))
			return
		}
		records := []*db_entity.DnsRecord{
			{
				RequestType: json.RequestType,
				Domain:      json.Domain,
				Address:     json.Address,
			},
		}
		err := srv.db.InsertRecord(records)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
			return
		}
		err = srv.cache.SetRecord(json.RequestType, json.Domain, json.Address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
			return
		}
		c.JSON(http.StatusOK, NewRespOK(nil))
	})
	dnsGroup.POST("/delete/:id", func(c *gin.Context) {
		var json struct {
			ID uint64 `json:"id" binding:"required"`
		}
		if err := c.ShouldBindUri(&json); err != nil {
			c.JSON(http.StatusBadRequest, NewRespFail(err.Error()))
			return
		}
		record, err := srv.db.GetRecordById(json.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
			return
		}
		err = srv.db.DeleteRecordById([]uint64{json.ID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
			return
		}
		err = srv.cache.DeleteRecord(record.RequestType, record.Domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
			return
		}
		c.JSON(http.StatusOK, NewRespOK(nil))
	})
	dnsGroup.POST("/delete", func(c *gin.Context) {
		var json struct {
			IDs []uint64 `json:"ids" binding:"required"`
		}
		if err := c.ShouldBind(&json); err != nil {
			c.JSON(http.StatusBadRequest, NewRespFail(err.Error()))
			return
		}
		dat, err := srv.db.GetRecordListById(json.IDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
			return
		}
		err = srv.db.DeleteRecordById(json.IDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
			return
		}
		for _, record := range dat {
			err = srv.cache.DeleteRecord(record.RequestType, record.Domain)
			if err != nil {
				c.JSON(http.StatusInternalServerError, NewRespFail(err.Error()))
				break
			}
		}
		c.JSON(http.StatusOK, NewRespOK(nil))
	})
	if err := r.Run(srv.cfg.Http.Listen); err != nil {
		return err
	}
	return nil
}
