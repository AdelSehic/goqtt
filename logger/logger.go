package logger

import (
	"goqtt/config"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Console zerolog.Logger
var HTTP zerolog.Logger

func Init(cfg *config.HttpLogger){
	Console = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.DateTime}).
	Level(zerolog.InfoLevel).
	With().Timestamp().Caller().Int("pid", os.Getpid()).Logger()

	HTTP = zerolog.New(&HttpLogger{
		Url: cfg.Url,
		Auth: "Bearer " + cfg.Auth,
		Client: &http.Client{},
	}).With().Timestamp().Caller().Logger()
}
