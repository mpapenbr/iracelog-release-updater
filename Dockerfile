FROM alpine:3.15
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
ENTRYPOINT ["/iracelog-release-updater"]
COPY iracelog-release-updater /
COPY config.yml /
EXPOSE 8000