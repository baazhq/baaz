FROM golang:1.20 as builder
ARG TARGETOS
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} go build -o manager cmd/main.go

FROM alpine:3.19
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/internal/chart internal/chart
ENTRYPOINT ["/manager"]
