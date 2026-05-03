# Stage 1: Build static assets with Node.js
FROM node:24-alpine AS builder

WORKDIR /build
COPY package.json package-lock.json ./
COPY webpack.config.js babel.config.json ./
COPY src/static/ ./src/static/
RUN npm install && npm run build

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS go-builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /sandwitches .

# Stage 3: Minimal runtime
FROM alpine:3.22
RUN apk add --no-cache ca-certificates supervisor
COPY --from=go-builder /sandwitches /app/sandwitches
COPY --from=builder /build/src/static/dist/ /app/static/dist/
COPY static/ /app/static/
COPY templates/ /app/templates/
COPY locale/ /app/locale/
COPY supervisord.conf /etc/supervisord.conf
WORKDIR /app
EXPOSE 6270
CMD ["supervisord", "-c", "/etc/supervisord.conf"]
