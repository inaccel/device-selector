FROM scratch
COPY device-selector /bin/device-selector
ENTRYPOINT ["device-selector"]
