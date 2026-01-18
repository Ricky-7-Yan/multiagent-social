FROM golang:1.24.11 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/multiagent ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/multiagent ./cmd/server
# NOTE: frontend build is performed by CI (web/react-app -> web/admin).
# Docker image build expects `web/admin` to already contain built static assets.

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /bin/multiagent /bin/multiagent
COPY --from=builder /src/web/admin /app/web/admin
EXPOSE 8080
ENTRYPOINT ["/bin/multiagent"]

