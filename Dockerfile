FROM scratch
ENTRYPOINT ["/iracelog-release-updater"]
COPY iracelog-release-updater /
COPY config.yml /