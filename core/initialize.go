package core

import (
	"context"
	"fmt"
	"gofound/global"
	"gofound/searcher"
	"gofound/searcher/words"
	"gofound/web/controller"
	"gofound/web/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewContainer(tokenizer *words.Tokenizer) *searcher.Container {
	container := &searcher.Container{
		Dir:       global.CONFIG.Data,
		Debug:     global.CONFIG.Debug,
		Tokenizer: tokenizer,
		Shard:     global.CONFIG.Shard,
		Timeout:   global.CONFIG.Timeout,
	}
	go container.Init()

	return container
}

func NewTokenizer(dictionaryPath string) *words.Tokenizer {
	return words.NewTokenizer(dictionaryPath)
}

// Initialize 初始化
func Initialize() {

	global.CONFIG = Parser()

	defer func() {

		if r := recover(); r != nil {
			fmt.Printf("panic: %s\n", r)
		}
	}()

	//初始化分词器
	tokenizer := NewTokenizer(global.CONFIG.Dictionary)
	global.Container = NewContainer(tokenizer)

	// 初始化业务逻辑
	controller.NewServices()

	// 注册路由
	r := router.SetupRouter()
	// 启动服务
	srv := &http.Server{
		Addr:    global.CONFIG.Addr,
		Handler: r,
	}
	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("listen:", err)
		}
	}()

	// 优雅关机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}

	log.Println("Server exiting")
}
