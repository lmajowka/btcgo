FROM golang
WORKDIR /app
RUN git clone -b main --single-branch --depth=1 https://github.com/lmajowka/btcgo.git btcgo
WORKDIR /app/btcgo
RUN rm -rf .git
RUN go mod tidy
RUN go build -o btcgo ./cmd/main.go
CMD ["./btcgo"]
