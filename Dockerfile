FROM golang:1.18 as base
RUN apt-get update && apt-get install -y ca-certificates openssl
# COPY MwCA.pem /etc/ssl/certs/MwCA.pem
RUN update-ca-certificates
ENV CGO_ENABLED=0
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /tmp/api-server ./*.go

FROM busybox:glibc as prod
RUN mkdir -p /var/log
COPY --from=base /etc/ssl/certs /etc/ssl/certs
COPY --from=base /tmp/api-server /usr/bin/api-server
COPY configyamls /app/configyamls
CMD ["api-server", "start"]