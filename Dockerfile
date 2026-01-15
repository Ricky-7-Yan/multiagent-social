FROM golang:1.24.11 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/multiagent ./cmd/server

# build frontend static assets into web/admin (optional - CI will run frontend build before docker build)
RUN if [ -d "web/react-app" ]; then \
  cd web/react-app && npm ci --silent && npm run build --silent && cd - && \
  rm -rf web/admin && mkdir -p web/admin && cp -r web/react-app/dist/* web/admin/ || true ; \
  fi

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /bin/multiagent /bin/multiagent
COPY --from=builder /src/web/admin /app/web/admin
EXPOSE 8080
ENTRYPOINT ["/bin/multiagent"]

