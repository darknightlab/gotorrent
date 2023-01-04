package main

import (
	"net"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/darknightlab/gotorrent/common"
	"golang.org/x/time/rate"
)

var (
	configPath string = "config/config.yaml"
)

func main() {
	var cfile ConfigFile
	ParseConfig(configPath, &cfile)
	var cfg Config
	// cfg.Web = struct {
	// 	Port    int
	// 	Address string
	// 	Secret  string
	// }{
	// 	Port:    16100,
	// 	Secret:  "",
	// 	Address: "",
	// }
	cfg.Web = cfile.Web
	// cfg.Main.CacheDir = cfile.Main.CacheDir
	// cfg.Main.MaxSeedTime = cfile.Main.MaxSeedTime
	// cfg.Main.GotInfoTimeout = cfile.Main.GotInfoTimeout
	// cfg.Main.SequentialDownload = cfile.Main.SequentialDownload
	// cfg.Main.CachePrefix = cfile.Main.CachePrefix
	// cfg.Main.DefaultTracker = cfile.Main.DefaultTracker
	cfg.Main = cfile.Main
	cfg.Engine = torrent.NewDefaultClientConfig()

	if cfile.Engine.DownloadRateLimit < 0 {
		cfg.Engine.DownloadRateLimiter = rate.NewLimiter(rate.Inf, 0)
	} else {
		cfg.Engine.DownloadRateLimiter = rate.NewLimiter(rate.Limit(cfile.Engine.DownloadRateLimit), cfile.Engine.DownloadRateBurst)
	}
	if cfile.Engine.UploadRateLimit < 0 {
		cfg.Engine.UploadRateLimiter = rate.NewLimiter(rate.Inf, 0)
	} else {
		cfg.Engine.UploadRateLimiter = rate.NewLimiter(rate.Limit(cfile.Engine.UploadRateLimit), cfile.Engine.UploadRateBurst)
	}
	cfg.Engine.ListenPort = cfile.Engine.ListenPort
	cfg.Engine.UpnpID = cfile.Engine.UpnpID
	cfg.Engine.ExtendedHandshakeClientVersion = cfile.Engine.ExtendedHandshakeClientVersion
	cfg.Engine.Bep20 = cfile.Engine.Bep20
	cfg.Engine.DataDir = cfile.Engine.DataDir
	cfg.Engine.Seed = cfile.Engine.Seed
	cfg.Engine.HTTPUserAgent = cfile.Engine.HTTPUserAgent
	cfg.Engine.DisableIPv6 = cfile.Engine.DisableIPv6
	cfg.Engine.PublicIp4 = net.ParseIP(cfile.Engine.PublicIp4)
	cfg.Engine.PublicIp6 = net.ParseIP(cfile.Engine.PublicIp6)
	// set default storage, saving sqlite in cache dir
	pc, err := storage.NewDefaultPieceCompletionForDir(cfg.Main.CacheDir)
	if err != nil {
		common.ClientPanic(err)
	}
	cfg.Engine.DefaultStorage = storage.NewFileWithCompletion(cfile.Engine.DataDir, pc)
	cl := New(cfg)
	cl.Listen()
}
