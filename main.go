package main

import (
	"github.com/ziyouzy/heartbeating"
)


func main(){
	defer heartbeating.LogFlush()
	heartbeating.Conf()

}