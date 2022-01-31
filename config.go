package main

import (
	"io/ioutil"

	"github.com/anacrolix/torrent"
	"github.com/darknightlab/gotorrent/common"
	"github.com/kataras/iris/v12"
	"gopkg.in/yaml.v3"
)

type WebConfig struct {
	Port    int    `yaml:"Port"`
	Address string `yaml:"Address"`
	Secret  string `yaml:"Secret"`
}

type MainConfig struct {
	CacheDir           string     `yaml:"CacheDir"`
	SequentialDownload bool       `yaml:"SequentialDownload"`
	CachePrefix        string     `yaml:"CachePrefix"`
	DefaultTracker     [][]string `yaml:"DefaultTracker"`
}

type ConfigFile struct {
	Web    WebConfig `yaml:"Web"`
	Engine struct {
		DownloadRateBurst              int     `yaml:"DownloadRateBurst"`
		DownloadRateLimit              float64 `yaml:"DownloadRateLimit"`
		UploadRateBurst                int     `yaml:"UploadRateBurst"`
		UploadRateLimit                float64 `yaml:"UploadRateLimit"`
		ListenPort                     int     `yaml:"ListenPort"`
		UpnpID                         string  `yaml:"UpnpID"`
		ExtendedHandshakeClientVersion string  `yaml:"ExtendedHandshakeClientVersion"`
		Bep20                          string  `yaml:"Bep20"`
		DataDir                        string  `yaml:"DataDir"`
		Seed                           bool    `yaml:"Seed"`
		HTTPUserAgent                  string  `yaml:"HTTPUserAgent"`
		DisableIPv6                    bool    `yaml:"DisableIPv6"`
		PublicIp4                      string  `yaml:"PublicIp4"`
		PublicIp6                      string  `yaml:"PublicIp6"`
	} `yaml:"Engine"`
	Main MainConfig `yaml:"Main"`
}

type Client struct {
	Engine *torrent.Client
	Web    *iris.Application
	Config Config
	Queue  struct {
		GotInfo map[string]*torrent.Torrent
	}
}

type Config struct {
	Engine *torrent.ClientConfig
	Web    WebConfig
	Main   MainConfig
}

func ParseConfig(configPath string, c *ConfigFile) {
	cFile, err := ioutil.ReadFile(configPath)
	common.ClientPanic(err)
	err = yaml.Unmarshal(cFile, c)
	common.ClientPanic(err)
}
