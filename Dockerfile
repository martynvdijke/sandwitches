# Stage 1: Build static assets with Node.js
FROM node:24-alpine AS builder

WORKDIR /build
COPY package.json package-lock.json ./
COPY webpack.config.js babel.config.json ./
COPY go-app/static/ ./go-app/static/
RUN npm install && npm run build

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS go-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY go-app/go.mod go-app/go.sum ./
RUN go mod download
COPY go-app/ ./
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /sandwitches .

# Stage 3: Minimal runtime
FROM alpine:3.22
RUN apk add --no-cache ca-certificates supervisor
COPY --from=go-builder /sandwitches /app/sandwitches
COPY --from=builder /build/go-app/static/dist/ /app/static/dist/
COPY go-app/static/ /app/static/
COPY go-app/templates/ /app/templates/
COPY go-app/supervisord.conf /etc/supervisord.conf
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
WORKDIR /app
EXPOSE 6270
CMD ["supervisord", "-c", "/etc/supervisord.conf"]
