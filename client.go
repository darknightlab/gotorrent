package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/darknightlab/gotorrent/common"
	"github.com/kataras/iris/v12"
)

func IsMagnet(uri string) bool {

	return (len(uri) > 20 && uri[:20] == "magnet:?xt=urn:btih:")
}

func GetMagnet(torr *torrent.Torrent) string {
	m := torr.Metainfo()
	return m.Magnet(nil, nil).String()
}

func GetConnStatusJSON(cs torrent.ConnStats) ConnStatusJSON {
	return ConnStatusJSON{
		BytesWritten:                cs.BytesWritten.Int64(),
		BytesWrittenData:            cs.BytesWrittenData.Int64(),
		BytesRead:                   cs.BytesRead.Int64(),
		BytesReadData:               cs.BytesReadData.Int64(),
		BytesReadUsefulData:         cs.BytesReadUsefulData.Int64(),
		BytesReadUsefulIntendedData: cs.BytesReadUsefulIntendedData.Int64(),
		ChunksWritten:               cs.ChunksWritten.Int64(),
		ChunksRead:                  cs.ChunksRead.Int64(),
		ChunksReadUseful:            cs.ChunksReadUseful.Int64(),
		ChunksReadWasted:            cs.ChunksReadWasted.Int64(),
		MetadataChunksRead:          cs.MetadataChunksRead.Int64(),
		PiecesDirtiedGood:           cs.PiecesDirtiedGood.Int64(),
		PiecesDirtiedBad:            cs.PiecesDirtiedBad.Int64(),
	}
}

func GetTorrentStatusJSON(t *torrent.Torrent) TorrentStatusJSON {
	ts := t.Stats()
	return TorrentStatusJSON{
		Name:             t.Name(),
		Hash:             t.InfoHash().String(),
		Magnet:           GetMagnet(t),
		ConnStatus:       GetConnStatusJSON(ts.ConnStats),
		TotalPeers:       ts.TotalPeers,
		PendingPeers:     ts.PendingPeers,
		ActivePeers:      ts.ActivePeers,
		ConnectedSeeders: ts.ConnectedSeeders,
		HalfOpenPeers:    ts.HalfOpenPeers,
		PiecesComplete:   ts.PiecesComplete,
	}
}

type ConnStatusJSON struct {
	// Total bytes on the wire. Includes handshakes and encryption.
	BytesWritten     int64 `json:"BytesWritten"`
	BytesWrittenData int64 `json:"BytesWrittenData"`

	BytesRead                   int64 `json:"BytesRead"`
	BytesReadData               int64 `json:"BytesReadData"`
	BytesReadUsefulData         int64 `json:"BytesReadUsefulData"`
	BytesReadUsefulIntendedData int64 `json:"BytesReadUsefulIntendedData"`

	ChunksWritten int64 `json:"ChunksWritten"`

	ChunksRead       int64 `json:"ChunksRead"`
	ChunksReadUseful int64 `json:"ChunksReadUseful"`
	ChunksReadWasted int64 `json:"ChunksReadWasted"`

	MetadataChunksRead int64 `json:"MetadataChunksRead"`

	// Number of pieces data was written to, that subsequently passed verification.
	PiecesDirtiedGood int64 `json:"PiecesDirtiedGood"`
	// Number of pieces data was written to, that subsequently failed verification. Note that a
	// connection may not have been the sole dirtier of a piece.
	PiecesDirtiedBad int64 `json:"PiecesDirtiedBad"`
}

type TorrentStatusJSON struct {
	Name   string `json:"Name"`
	Hash   string `json:"Hash"`
	Magnet string `json:"Magnet"`
	// Aggregates stats over all connections past and present. Some values may not have much meaning
	// in the aggregate context.
	ConnStatus ConnStatusJSON `json:"ConnStatus"`

	// Ordered by expected descending quantities (if all is well).
	TotalPeers       int `json:"TotalPeers"`
	PendingPeers     int `json:"PendingPeers"`
	ActivePeers      int `json:"ActivePeers"`
	ConnectedSeeders int `json:"ConnectedSeeders"`
	HalfOpenPeers    int `json:"HalfOpenPeers"`
	PiecesComplete   int `json:"PiecesComplete"`
}

type Auth struct {
	Secret string `json:"Secret"`
}

// <Local Method

func (cl *Client) NewTorrentCache(t *torrent.Torrent) {
	d, err := os.Stat(cl.Config.Main.CacheDir)
	common.ClientPanic(err)
	if d.IsDir() {
		cacheFilePath := filepath.Join(cl.Config.Main.CacheDir, fmt.Sprintf("%s%s.torrent", cl.Config.Main.CachePrefix, t.InfoHash().String()))
		f, err := os.Stat(cacheFilePath)
		if os.IsNotExist(err) {
			f, err := os.Create(cacheFilePath)
			common.ClientPanic(err)
			defer f.Close()
			t.Metainfo().Write(f)
			common.ClientInfo(fmt.Sprintf("created cache file: %s", cacheFilePath))
		} else if err == nil {
			if f.IsDir() {
				common.ClientPanic(errors.New("filepath is a dir"))
			}
		} else {
			common.ClientPanic(err)
		}
	}
}

func (cl *Client) Recover() {
	d, _ := ioutil.ReadDir(cl.Config.Main.CacheDir)
	for _, f := range d {
		if !f.IsDir() {
			torr, err := cl.Engine.AddTorrentFromFile(filepath.Join(cl.Config.Main.CacheDir, f.Name()))
			if err != nil {
				common.ClientError(fmt.Errorf("add %s", f.Name()))
				continue
			}
			if f.Name() != fmt.Sprintf("%s%s.torrent", cl.Config.Main.CachePrefix, torr.InfoHash().String()) {
				common.ClientError(fmt.Errorf("%s hash dismatch", f.Name()))
			}
			torr.DownloadAll()
			common.ClientInfo(fmt.Sprintf("add %s", f.Name()))
		}
	}
}

func (cl *Client) AddTorrentFromFile(file io.Reader) (*torrent.Torrent, error) {
	m, err := metainfo.Load(file)
	if err != nil {
		return nil, err
	}
	torr, err := cl.Engine.AddTorrent(m)
	if err != nil {
		return nil, err
	}
	return torr, nil
}

func (cl *Client) DownloadTorrent(torr *torrent.Torrent) {
	torr.AddTrackers(cl.Config.Main.DefaultTracker)

	cl.NewTorrentCache(torr)
	torr.DownloadAll()

	// 大于8M优先下载前后部分，方便读出视频时长
	pieceLength := torr.Info().PieceLength
	var minFileSize int64 = 16777216
	for i := 0; i < len(torr.Files()); i++ {
		file := torr.Files()[i]
		if file.Length() > minFileSize {
			begin := file.BeginPieceIndex()
			end := file.EndPieceIndex()
			n := int(minFileSize / 2 / pieceLength) // 潜在的不安全行为：int64->int
			for p := 0; p < n; p++ {
				torr.Piece(begin + p).SetPriority(torrent.PiecePriorityNow)
				torr.Piece(end - 1 - p).SetPriority(torrent.PiecePriorityNow)
			}
		}
	}

	if cl.Config.Main.SequentialDownload {
		go func() {
			r := torr.NewReader()
			io.Copy(ioutil.Discard, r)
		}()
	}

	common.ClientInfo(fmt.Sprintf("start download %s", torr.Name()))
}

func (cl *Client) DeleteTorrent(hash string, deleteFile bool) {
	t, ok := cl.Engine.Torrent(metainfo.NewHashFromHex(hash))
	if ok {
		t.Drop()
		os.Remove(filepath.Join(cl.Config.Main.CacheDir, fmt.Sprintf("%s%s.torrent", cl.Config.Main.CachePrefix, t.InfoHash().String())))
		common.ClientInfo(fmt.Sprintf("removed cache file: %s", fmt.Sprintf("%s%s.torrent", cl.Config.Main.CachePrefix, t.InfoHash().String())))
		if deleteFile {
			os.RemoveAll(filepath.Join(cl.Config.Engine.DataDir, t.Name()))
			common.ClientInfo(fmt.Sprintf("removed torrent files: %s", t.Name()))
		}
	} else {
		common.UserPanic(fmt.Sprintf("torrent not exist: %s", hash))
		return
	}
}

func (cl *Client) Close(ctx iris.Context) {

}

func (cl *Client) Closed(ctx iris.Context) {
	<-cl.Engine.Closed()
	ctx.JSON(map[string]interface{}{"response": 200, "info": "成功关闭"})
}

func (cl *Client) CheckSecret(secret string) bool {
	return secret == cl.Config.Web.Secret
}

func (cl *Client) ConnStatus() ConnStatusJSON {
	return GetConnStatusJSON(cl.Engine.ConnStats())
}

func (cl *Client) TorrentStatus(torr *torrent.Torrent) TorrentStatusJSON {
	return GetTorrentStatusJSON(torr)
}

func (cl *Client) TorrentsStatus() []TorrentStatusJSON {
	var res []TorrentStatusJSON
	for _, t := range cl.Engine.Torrents() {
		res = append(res, cl.TorrentStatus(t))
	}
	return res
}

func (cl *Client) StopTorrent(torr *torrent.Torrent) {
	torr.DisallowDataDownload()
	torr.DisallowDataUpload()
}

func (cl *Client) StartTorrent(torr *torrent.Torrent) {
	torr.AllowDataDownload()
	torr.AllowDataUpload()
}

// Local Method>

func (cl *Client) Listen() {
	cl.Web.Listen(fmt.Sprintf("%s:%d", cl.Config.Web.Address, cl.Config.Web.Port))
}

func New(cfg Config) (client *Client) {
	client = new(Client)
	client.Config = cfg
	// client.Config.CacheDir, _ = filepath.Abs(client.Config.CacheDir)
	// client.Config.Engine.DataDir, _ = filepath.Abs(client.Config.Engine.DataDir)
	client.Queue.GotInfo = map[string]*torrent.Torrent{}

	{
		d, err := os.Stat(client.Config.Main.CacheDir)
		if os.IsNotExist(err) {
			os.Mkdir(client.Config.Main.CacheDir, os.ModePerm)
		} else if !d.IsDir() {
			common.ClientPanic(errors.New("cachedir is not a dir"))
		}
		d, err = os.Stat(client.Config.Engine.DataDir)
		if os.IsNotExist(err) {
			os.Mkdir(client.Config.Engine.DataDir, os.ModePerm)
		} else if !d.IsDir() {
			common.ClientPanic(errors.New("datadir is not a dir"))
		}
	}

	client.Engine, _ = torrent.NewClient(client.Config.Engine)

	client.Recover()

	client.Web = iris.New()

	// CORS
	// crs := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"*"},
	// 	AllowCredentials: true,
	// 	AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	// 	AllowedHeaders:   []string{"Access-Control-Allow-Origin", "Content-Type"},
	// 	MaxAge:           86400,
	// })
	crs := func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Methods",
				"POST, PUT, PATCH, DELETE, GET")

			ctx.Header("Access-Control-Allow-Headers",
				"Access-Control-Allow-Origin,Content-Type")

			ctx.Header("Access-Control-Max-Age",
				"86400")

			ctx.StatusCode(iris.StatusNoContent)
			return
		}

		ctx.Next()
	}
	client.Web.Use(crs)

	client.Web.HandleDir("/", "./html")
	client.Web.HandleDir("/webseed/", cfg.Engine.DataDir)
	api := client.Web.Party("/api")
	{
		add := api.Party("/add")
		{
			add.Post("/uri", client.WebAddURI)
			add.Post("/torrent", client.WebAddTorrentFromFile)
		}

		close := api.Party("/close")
		{
			close.Post("/", client.Closed)
		}

		// 删除某一个torrent，要求torrent序号
		delete := api.Party("/delete")
		{
			delete.Post("/", client.WebDelete)
		}

		// 整个客户端的状态
		status := api.Party("/status")
		{
			status.Post("/", client.WebStatus)
			status.Get("/", client.WebStatus)
		}

		torr := api.Party("/torrent")
		{
			torr.Post("/{hash:string}/delete", client.WebDeleteTorrent)
			torr.Post("/{hash:string}/status", client.WebTorrentStatus)
			torr.Get("/{hash:string}/status", client.WebTorrentStatus)
		}
	}
	return
}

func (cl *Client) WebAddURI(ctx iris.Context) {
	var recv struct {
		Auth `json:"Auth"`
		URI  string `json:"URI"`
	}
	ctx.ReadJSON(&recv)
	if !cl.CheckSecret(recv.Auth.Secret) {
		ctx.JSON(map[string]interface{}{"response": 401, "info": "Secret错误"})
		return
	}

	var torr *torrent.Torrent
	if IsMagnet(recv.URI) {
		var err error
		torr, err = cl.Engine.AddMagnet(recv.URI)
		if err != nil {
			ctx.JSON(map[string]interface{}{"response": 400, "info": err.Error()})
			common.UserPanic(err.Error())
			return
		}
	} else {
		resp, err := http.Get(recv.URI)
		if err != nil {
			ctx.JSON(map[string]interface{}{"response": 400, "info": err.Error()})
			return
		}
		defer resp.Body.Close()
		torr, err = cl.AddTorrentFromFile(resp.Body)
		if err != nil {
			ctx.JSON(map[string]interface{}{"response": 400, "info": err.Error()})
			common.UserPanic(err.Error())
			return
		}
	}

	common.ClientInfo(fmt.Sprintf("receive uri: %s", recv.URI))

	cl.Queue.GotInfo[torr.InfoHash().String()] = torr
	<-torr.GotInfo()
	delete(cl.Queue.GotInfo, torr.InfoHash().String())

	cl.DownloadTorrent(torr)
	ctx.JSON(map[string]interface{}{"response": 200, "magnet": GetMagnet(torr)})
}

func (cl *Client) WebAddTorrentFromFile(ctx iris.Context) {
	torrFile, _, err := ctx.FormFile("torrent")
	if err != nil {
		ctx.JSON(map[string]interface{}{"response": 400, "info": err.Error()})
		common.UserPanic(err.Error())
		return
	}
	torr, err := cl.AddTorrentFromFile(torrFile)
	if err != nil {
		ctx.JSON(map[string]interface{}{"response": 400, "info": err.Error()})
		common.UserPanic(err.Error())
		return
	}

	common.ClientInfo(fmt.Sprintf("receive torrent: %s", torr.Name()))
	cl.DownloadTorrent(torr)
	ctx.JSON(map[string]interface{}{"response": 200, "magnet": GetMagnet(torr)})
}

func (cl *Client) WebDelete(ctx iris.Context) {
	var recv struct {
		Auth       `json:"Auth"`
		Hash       string `json:"Hash"`
		DeleteFile string `json:"DeleteFile"`
	}
	ctx.ReadJSON(&recv)
	if !cl.CheckSecret(recv.Auth.Secret) {
		ctx.JSON(map[string]interface{}{"response": 401, "info": "Secret错误"})
		common.UserPanic("Secret错误")
		return
	}
	cl.DeleteTorrent(recv.Hash, recv.DeleteFile == "yes")
	ctx.JSON(map[string]interface{}{"response": 200, "info": "删除成功"})
}

func (cl *Client) WebDeleteTorrent(ctx iris.Context) {
	var recv struct {
		Auth       `json:"Auth"`
		DeleteFile string `json:"DeleteFile"`
	}
	ctx.ReadJSON(&recv)
	if !cl.CheckSecret(recv.Auth.Secret) {
		ctx.JSON(map[string]interface{}{"response": 401, "info": "Secret错误"})
		common.UserPanic("Secret错误")
		return
	}
	cl.DeleteTorrent(ctx.Params().Get("hash"), recv.DeleteFile == "yes")
	ctx.JSON(map[string]interface{}{"response": 200, "info": "删除成功"})
}

func (cl *Client) WebStatus(ctx iris.Context) {
	// var recv struct {
	// 	Auth `json:"Auth"`
	// }
	// ctx.ReadJSON(&recv)
	// if !cl.CheckSecret(recv.Auth.Secret) {
	// 	ctx.JSON(map[string]interface{}{"response": 401, "info": "Secret错误"})
	// 	common.UserPanic("Secret错误")
	// 	return
	// }
	resp := struct {
		Total    ConnStatusJSON      `json:"Total"`
		Torrents []TorrentStatusJSON `json:"Torrents"`
	}{
		Total:    cl.ConnStatus(),
		Torrents: cl.TorrentsStatus(),
	}
	ctx.JSON(resp)
}

func (cl *Client) WebStartTorrent(ctx iris.Context) {
	t, ok := cl.Engine.Torrent(metainfo.NewHashFromHex(ctx.Params().Get("hash")))
	if ok {
		cl.StartTorrent(t)
		ctx.JSON(map[string]interface{}{"response": 200, "info": "成功"})
	} else {
		ctx.JSON(map[string]interface{}{"response": 400, "info": "没有找到torrent"})
	}
}

func (cl *Client) WebStopTorrent(ctx iris.Context) {
	t, ok := cl.Engine.Torrent(metainfo.NewHashFromHex(ctx.Params().Get("hash")))
	if ok {
		cl.StopTorrent(t)
		ctx.JSON(map[string]interface{}{"response": 200, "info": "成功"})
	} else {
		ctx.JSON(map[string]interface{}{"response": 400, "info": "没有找到torrent"})
	}
}

func (cl *Client) WebTorrentStatus(ctx iris.Context) {
	// var recv struct {
	// 	Auth `json:"Auth"`
	// }
	// ctx.ReadJSON(&recv)
	// if !cl.CheckSecret(recv.Auth.Secret) {
	// 	ctx.JSON(map[string]interface{}{"response": 401, "info": "Secret错误"})
	// 	return
	// }

	t, ok := cl.Engine.Torrent(metainfo.NewHashFromHex(ctx.Params().Get("hash")))
	if ok {
		ctx.JSON(cl.TorrentStatus(t))
	} else {
		ctx.JSON(map[string]interface{}{"response": 400, "info": "没有找到torrent"})
	}
}
