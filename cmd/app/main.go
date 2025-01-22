package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gopkg.in/fsnotify.v1"

	"github.com/smeshkov/gomock/app"
	"github.com/smeshkov/gomock/config"
)

var (
	version = "untagged"
)

func main() {
	configFile := flag.String("config", "_resources/config.yml", "Configuration file")
	mockFile := flag.String("mock", "mock.json", "Mock configuration file")
	verbose := flag.Bool("verbose", false, "Verbose")
	ver := flag.Bool("version", false, "prints version of Tagify")
	watch := flag.Bool("watch", false, "Watch config file changes and reload automatically")

	flag.Parse()

	if *ver {
		fmt.Println(version)
		return
	}

	// 초기 로깅 설정
	config.SetupLog("info")

	// 종료 신호를 처리할 채널
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 서버 실행 루프
	for {
		// 초기 설정 로드
		cfg, err := config.NewConfig(*configFile)
		if err != nil {
			zap.L().Warn(fmt.Sprintf("failed to load configuration %s: %v", *configFile, err))
		}

		logLevel := cfg.Logger.Level
		if *verbose {
			logLevel = "debug"
		}

		config.SetupLog(logLevel)

		var mck config.Mock
		var mockPath string
		mck, mockPath, err = config.NewMock(*mockFile)
		if err != nil {
			zap.L().Warn(fmt.Sprintf("failed to load API configuration %s: %v", *mockFile, err))
		}

		// 서버 실행 및 감시
		serverCtx, cancelServer := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			runServer(serverCtx, &cfg, &mck, version, mockPath)
		}()

		// 설정 변경 감시 활성화
		if *watch {
			go watchConfigFiles(*configFile, *mockFile, cancelServer)
		}

		select {
		case <-sigChan:
			// 종료 신호 수신 시 서버 종료
			zap.L().Info("received termination signal, shutting down...")
			cancelServer()
			wg.Wait()
			return
		case <-serverCtx.Done():
			// 설정 변경으로 서버가 재시작될 때
			zap.L().Info("restarting server due to configuration change...")
			wg.Wait()
			// 루프를 계속 돌면서 새로운 설정을 반영한 서버 실행
		}
	}
}

// 서버 실행 함수
func runServer(ctx context.Context, cfg *config.Config, mck *config.Mock, version, mockPath string) {
	srv := &http.Server{
		ReadHeaderTimeout: cfg.Server.ReadTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		Addr:              cfg.Server.Addr,
		Handler:           app.RegisterHandlers(version, mockPath, cfg, mck),
	}

	go func() {
		zap.L().Info(fmt.Sprintf("starting proxy app on %s (read timeout %s, write timeout %s)",
			cfg.Server.Addr, cfg.Server.ReadTimeout.String(), cfg.Server.WriteTimeout.String()))

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.L().Fatal(fmt.Sprintf("failed to start server: %v", err))
		}
	}()

	<-ctx.Done()
	zap.L().Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Fatal(fmt.Sprintf("server shutdown failed: %v", err))
	}
}

// 설정 파일 변경 감지 및 서버 재시작
func watchConfigFiles(configPath, mockPath string, cancelServer context.CancelFunc) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.L().Fatal(fmt.Sprintf("failed to create watcher: %v", err))
	}
	defer watcher.Close()

	// 감시할 파일 추가
	files := []string{configPath, mockPath}
	for _, file := range files {
		err = watcher.Add(file)
		if err != nil {
			zap.L().Fatal(fmt.Sprintf("failed to watch file %s: %v", file, err))
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				zap.L().Info(fmt.Sprintf("file %s changed, restarting server...", event.Name))
				cancelServer() // 서버 종료 트리거
				return         // 감시 루프 종료, main 루프에서 서버 재시작
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			zap.L().Error(fmt.Sprintf("watcher error: %v", err))
		}
	}
}
