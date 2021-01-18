package heartbeating

import (
	"heartbeating/conf"
	"heartbeating/log"
)


func HeartBeating(){
	defer log.LogFlush()
	conf.Conf()

}