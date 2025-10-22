FROM golang:1.25.1 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make bin

FROM scratch AS release
COPY --from=builder /app/tmp/bin/kart /app
ENTRYPOINT [ "/app" ]
