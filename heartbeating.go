package main

import (
	"heartbeating/conf"
	"heartbeating/log"
)


func main(){
	defer log.LogFlush()
	conf.Conf()

}