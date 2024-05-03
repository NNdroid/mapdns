package main

import (
	"flag"
	"golang.org/x/net/context"
	cache_srv "mapdns/pkg/cache-srv"
	"mapdns/pkg/common"
	"mapdns/pkg/config"
	"mapdns/pkg/db-srv"
	"mapdns/pkg/dns-srv"
	"mapdns/pkg/http-srv"
	"mapdns/pkg/log"
	"os"
	"sync"
)

var (
	_configFilePath string
	_flagQuiet      bool
	_config         *config.Config
	_flagVersion    bool
)

func init() {
	flag.StringVar(&_configFilePath, "c", "config.yaml", "the path of configuration file")
	flag.BoolVar(&_flagQuiet, "quiet", false, "quiet for log print.")
	flag.BoolVar(&_flagVersion, "v", false, "print version info.")
	flag.Parse()
	if _flagVersion {
		common.PrintVersion()
		os.Exit(0)
	}
	log.SetVerbose(!_flagQuiet)
	if !common.IsFile(_configFilePath) || !common.ExistsFile(_configFilePath) {
		log.Logger().Fatal("configure file not found!")
	}
	var err error
	_config, err = config.ReadConfig(_configFilePath)
	if err != nil {
		log.Logger().Fatalf("failed to read configuration file: %v", err)
	}
	_config.Verbose = !_flagQuiet
}

type App struct {
	ctx   context.Context
	wg    *sync.WaitGroup
	http  *http_srv.Server
	dns   *dns_srv.Server
	db    *db_srv.Server
	cache *cache_srv.Server
}

func NewApp(ctx context.Context) *App {
	return &App{ctx: ctx, wg: new(sync.WaitGroup)}
}

func (app *App) runHttpServer() {
	srv := http_srv.New(app.ctx, _config, app.db, app.cache)
	if err := srv.ListenAndServe(); err != nil {
		app.wg.Done()
		log.Logger().Fatalf("http server failed: %v", err)
	}
}

func (app *App) runDnsServer() {
	srv := dns_srv.New(app.ctx, _config, app.db, app.cache)
	log.Logger().Infof("starting dns-srv at %s", _config.DNS.Listen)
	if err := srv.ListenAndServe(); err != nil {
		app.wg.Done()
		log.Logger().Fatalf("failed to start dns-srv: %v", err)
	}
}

func (app *App) runDbServer() {
	app.db = db_srv.New(app.ctx, _config)
}

func (app *App) runCacheServer() {
	app.cache = cache_srv.New(app.ctx, _config)
}

func main() {
	app := NewApp(context.TODO())
	app.runDbServer()
	app.runCacheServer()
	app.wg.Add(2)
	go app.runHttpServer()
	go app.runDnsServer()
	app.wg.Wait()
}
