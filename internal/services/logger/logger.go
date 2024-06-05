package logger

import (
	"os"
	"sync"
	"time"
	"websocket-confee/internal/models"
)

const Debug = "DEBUG"
const Info = "INFO"
const Warn = "WARN"
const Error = "ERROR"

type logInterface interface {
	BatchSave(logs map[string]models.Log)
}

const batchSize = 250

type Logger struct {
	logger    logInterface
	logsBatch map[string]models.Log
	sync.RWMutex
}

func NewLogger() *Logger {
	logsBatch := make(map[string]models.Log, batchSize+10)

	if os.Getenv("APP_ENV") == "prod" {
		logger, err := newClickhouseLogger()
		if err != nil {
			panic("cant init ch logger")
		}

		return &Logger{logger: logger, logsBatch: logsBatch}
	}

	logger, err := newFileLogger()
	if err != nil {
		panic("cant init default logger")
	}

	return &Logger{logger: logger, logsBatch: logsBatch}
}
func (c *Logger) Debug(msg string) {
	c.checkLogMap()
	c.addLog(Debug, msg)

}
func (c *Logger) Info(msg string) {
	c.checkLogMap()
	c.addLog(Info, msg)
}
func (c *Logger) Warn(msg string) {
	c.checkLogMap()
	c.addLog(Warn, msg)
}
func (c *Logger) Error(msg string) {
	c.checkLogMap()
	c.addLog(Error, msg)
}

func (c *Logger) BatchSave() {
	c.logger.BatchSave(c.logsBatch)
}

func (c *Logger) addLog(logLevel string, msg string) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	c.logsBatch[logLevel] = models.Log{
		Msg:  msg,
		Time: time.Now(),
	}
}

func (c *Logger) checkLogMap() {
	if len(c.logsBatch) >= batchSize {
		c.BatchSave()
	}
}
