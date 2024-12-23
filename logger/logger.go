package logger

import (
	"goqtt/config"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Default zerolog.Logger
var HTTP zerolog.Logger

func Init(cfg *config.HttpLogger) {
	logFile, err := os.OpenFile("goqtt.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	fileWriter := zerolog.ConsoleWriter{
		Out:        logFile,
		TimeFormat: time.RFC3339,
	}

	Default = zerolog.New(fileWriter).Level(zerolog.InfoLevel).
		With().Timestamp().Caller().Logger()

	HTTP = zerolog.New(&HttpLogger{
		Url:    cfg.Url,
		Auth:   "Bearer " + cfg.Auth,
		Client: &http.Client{},
	}).With().Timestamp().Caller().Logger()
}
