# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git make gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s -extldflags "-static"' -o bin/manager cmd/manager/main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workspace/bin/manager .
COPY --from=builder /workspace/config/samples config/samples

USER nonroot:nonroot

ENTRYPOINT ["/manager"]
