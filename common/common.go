package common

import (
	"log"
	"os"
)

const Version = "0.1.3"
const PeerIDPrefix = "-GT0130-"

func PrintVersion() {
	log.Printf("gotorrent v%s", Version)
}

func UserError(info string) {
	log.Printf("[Info] [User:Error]%s", info)
}

func UserPanic(info string) {
	log.Printf("[Info] [User:Fatal]%s", info)
}

func ClientError(err error) {
	if err != nil {
		log.Printf("[Error] %s", err.Error())
	}
}

func ClientPanic(err error) {
	if err != nil {
		log.Panicf("[Fatal] %s", err.Error())
	}
}

func ClientInfo(info string) {
	log.Printf("[Info]  %s", info)
}

// 判断文件夹是否存在
func IsDirExist(path string) bool {
	s, err := os.Stat(path)
	return (err == nil && s.IsDir()) || os.IsExist(err)
}

// 判断文件是否存在
func IsFileExist(path string) bool {
	s, err := os.Stat(path)
	return (err == nil && !s.IsDir()) || os.IsExist(err)
}

// 创建文件夹
func CreateDir(path string) error {
	if !IsDirExist(path) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}
