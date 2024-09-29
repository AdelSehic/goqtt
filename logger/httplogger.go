package logger

import (
	"bytes"
	"fmt"
	"net/http"
)

type HttpLogger struct {
	Url string
	Auth string
	Client *http.Client
}

func (sw *HttpLogger) Write(p []byte) (n int, err error){
	r, err := http.NewRequest("POST", sw.Url, bytes.NewBuffer(p))
	if err != nil {
		Console.Err(err).Msg("Failed to create Logger POST request body")
		fmt.Println()
		return 0, err
	}
	r.Header.Add("Contnt-Type", "application/json")
	r.Header.Add("Authorization", "Bearer " + sw.Auth)

	resp, err := sw.Client.Do(r)
	if err != nil {
		Console.Err(err).Msg("Failed to send Logger POST request")
		return 0, err
	}
	resp.Body.Close()

	return 0, nil
}
