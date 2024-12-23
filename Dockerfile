FROM golang:1.23.1

WORKDIR /app

COPY . /app

RUN go mod download

WORKDIR cmd
RUN go build -o goqtt .

EXPOSE 8080

ENV CONFIG_FILE /app/config/config.json
cmd ["./goqtt"]
