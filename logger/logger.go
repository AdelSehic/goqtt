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
	logWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	Default = zerolog.New(logWriter).Level(zerolog.InfoLevel).
		With().Timestamp().Caller().Logger()

	HTTP = zerolog.New(&HttpLogger{
		Url:    cfg.Url,
		Auth:   "Bearer " + cfg.Auth,
		Client: &http.Client{},
	}).With().Timestamp().Caller().Logger()
}
