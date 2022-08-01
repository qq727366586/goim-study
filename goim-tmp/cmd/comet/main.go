package main

import (
	"flag"
	"goim-study/goim-tmp/internal/comet/conf"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 初始化server

}
