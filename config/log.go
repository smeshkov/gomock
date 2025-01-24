package config

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var (
	logger     *zap.Logger
	loggerLock sync.Mutex
)

// SetupLog 설정 함수 (초기화 시 1회 실행, 이후 레벨만 변경)
func SetupLog(level string) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	if logger == nil {
		// 처음 실행할 때만 로거 생성
		newLogger, err := newLog(level)
		if err != nil {
			fmt.Printf("failed to setup logger: %v\n", err)
			return
		}
		logger = newLogger
		zap.ReplaceGlobals(logger) // 전역 로거 설정
		zap.L().Sugar().Infof("logger initialized with level [%s]", level)
	} else {
		// 기존 로거가 있으면 레벨만 변경
		zap.L().Sugar().Infof("updating logger level to [%s]", level)
		_ = logger.Sync() // 기존 로거 버퍼 플러시
		newLogger, _ := newLog(level)
		logger = newLogger
	}
}

// SyncLog 버퍼 플러시 함수
func SyncLog() {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	if logger != nil {
		_ = logger.Sync()
	}
}

// newLog 내부 함수 (로거 생성)
func newLog(level string) (*zap.Logger, error) {
	if level != "info" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
