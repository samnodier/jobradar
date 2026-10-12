FROM node:24-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25.8-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
ARG REVISION=development
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-X main.buildRevision=${REVISION}" -o /jobradar ./cmd/jobradar

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 jobradar
WORKDIR /app
COPY --from=backend /jobradar /app/jobradar
COPY --from=frontend /app/frontend/dist /app/frontend/dist
USER jobradar
EXPOSE 8080
ENTRYPOINT ["/app/jobradar"]
