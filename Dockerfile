FROM golang AS stage1
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o btcgo ./cmd/main.go

FROM scratch
COPY --from=stage1 /app/btcgo /
COPY --from=stage1 /app/data /data
COPY --from=stage1 /app/.env /.env
CMD ["./btcgo"]
ENTRYPOINT ["/btcgo"]
