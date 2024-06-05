package logger

import (
	"log"
	"os"
	"time"
	"websocket-confee/internal/models"
)

type FileLogger struct {
	logger *log.Logger
}

func newFileLogger() (*FileLogger, error) {
	logFile, err := os.OpenFile("dev.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	return &FileLogger{
		logger: log.New(logFile, "LOGGER ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}
func (f *FileLogger) Debug(msg string) {
	f.logger.Printf("[%s] msg=%s time=%s", Debug, msg, time.Now().String())
}
func (f *FileLogger) Info(msg string) {
	f.logger.Printf("[%s] msg=%s time=%s", Info, msg, time.Now().String())
}
func (f *FileLogger) Warn(msg string) {
	f.logger.Printf("[%s] msg=%s time=%s", Warn, msg, time.Now().String())
}
func (f *FileLogger) Error(msg string) {
	f.logger.Printf("[%s] msg=%s time=%s", Error, msg, time.Now().String())
}
func (f *FileLogger) BatchSave(logs map[string]models.Log) {
	for level, logStruct := range logs {
		f.logger.Printf("[%s] msg=%s time=%s", level, logStruct.Msg, logStruct.Time.String())
	}
}
