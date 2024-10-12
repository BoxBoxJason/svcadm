/*
package logger is the package that contains the logger for the application.

The logger is used to log the application's activity to the console and to a file.
*/

package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/boxboxjason/svcadm/internal/static"
)

// Logger instances
var (
	LOG_LEVEL      int
	mu_LOG_LEVEL   sync.RWMutex
	debugLogger    *log.Logger
	infoLogger     *log.Logger
	errorLogger    *log.Logger
	criticalLogger *log.Logger
	fatalLogger    *log.Logger
	log_levels     = map[string]int{
		"DEBUG":    0,
		"INFO":     1,
		"ERROR":    2,
		"CRITICAL": 3,
		"FATAL":    4,
	}
)

// SetLogLevel sets the log level
func SetLogLevel(level string) error {
	mu_LOG_LEVEL.Lock()
	defer mu_LOG_LEVEL.Unlock()
	if _, ok := log_levels[strings.ToUpper(level)]; !ok {
		return errors.New("invalid log level " + level)
	}
	LOG_LEVEL = log_levels[strings.ToUpper(level)]
	return nil
}

// GetLogLevel gets the log level
func GetLogLevel() int {
	mu_LOG_LEVEL.RLock()
	defer mu_LOG_LEVEL.RUnlock()
	return LOG_LEVEL
}

func init() {
	// Check if the directory for the log file exists
	LOG_DIR := fmt.Sprintf("%s/log", static.SVCADM_HOME)
	LOG_FILE := fmt.Sprintf("%s/svcadm.log", LOG_DIR)
	if _, err := os.Stat(LOG_DIR); os.IsNotExist(err) {
		err = os.MkdirAll(LOG_DIR, os.ModePerm)
		if err != nil {
			log.Fatalln("Failed to create log directory:", err)
		}
	}

	file, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}

	multi := io.MultiWriter(file, os.Stdout)

	// Initialize loggers
	debugLogger = log.New(multi, "DEBUG: ", log.Ldate|log.Ltime)
	infoLogger = log.New(multi, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(multi, "ERROR: ", log.Ldate|log.Ltime)
	criticalLogger = log.New(multi, "CRITICAL: ", log.Ldate|log.Ltime)
	fatalLogger = log.New(multi, "FATAL: ", log.Ldate|log.Ltime)
}

// Debug logs a debug message.
func Debug(v ...interface{}) {
	if LOG_LEVEL <= 0 {
		debugLogger.Println(v...)
	}
}

// Info logs an info message.
func Info(v ...interface{}) {
	if LOG_LEVEL <= 1 {
		infoLogger.Println(v...)
	}
}

// Error logs an error message.
func Error(v ...interface{}) {
	if LOG_LEVEL <= 2 {
		errorLogger.Println(v...)
	}
}

// Critical logs a critical message but does not exit the application.
func Critical(v ...interface{}) {
	if LOG_LEVEL <= 3 {
		criticalLogger.Println(v...)
	}
}

// Fatal logs a fatal message and exits the application.
func Fatal(v ...interface{}) {
	if LOG_LEVEL <= 4 {
		fatalLogger.Fatalln(v...)
	}
}
