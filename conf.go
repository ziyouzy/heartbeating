package heartbeating

import (
	logger "github.com/phachon/go-logger"
)


func Conf(lp *logger.Logger){
	Logger(lp)
}


func ConfFlush(){
	LogFlush()
}