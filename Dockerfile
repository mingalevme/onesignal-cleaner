FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .

FROM alpine AS app
COPY --from=builder /app/app /usr/local/bin/app
ENTRYPOINT [ "/usr/local/bin/app", "onesignal-cleaner" ]
