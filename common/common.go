package common

import (
	"fmt"
	"log"
)

func UserError(info string) {
	log.Println(fmt.Sprintf("[Info] [User:Error]%s", info))
}

func UserPanic(info string) {
	log.Println(fmt.Sprintf("[Info] [User:Fatal]%s", info))
}

func ClientError(err error) {
	if err != nil {
		log.Println(fmt.Sprintf("[Error] %s", err.Error()))
	}
}

func ClientPanic(err error) {
	if err != nil {
		log.Panicln(fmt.Sprintf("[Fatal] %s", err.Error()))
	}
}

func ClientInfo(info string) {
	log.Println(fmt.Sprintf("[Info]  %s", info))
}
