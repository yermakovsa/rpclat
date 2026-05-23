FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/rpclat ./cmd/rpclat

FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
	&& adduser -D -H -s /sbin/nologin rpclat

COPY --from=build /out/rpclat /usr/local/bin/rpclat

USER rpclat

ENTRYPOINT ["rpclat"]