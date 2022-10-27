FROM golang:1.18 as base
RUN apt-get update && apt-get install -y ca-certificates openssl
COPY MwCA.pem /etc/ssl/certs/MwCA.pem
RUN update-ca-certificates

FROM base as local

WORKDIR /usr/
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s
WORKDIR /app

RUN echo "#!/bin/bash\nair --build.cmd \"go build -o /tmp/api-server /app/*.go\" --build.bin \"/tmp/api-server \$*\""> /usr/bin/api-server && chmod +x /usr/bin/api-server



FROM base as build
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0

RUN go get -d -v ./... && go mod tidy
RUN CGO_ENABLED=0 go build -o /tmp/api-server ./*.go

FROM busybox:glibc as prod
RUN mkdir -p /var/log
WORKDIR /app
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /tmp/api-server /usr/bin/api-server
COPY --from=build /app/configyamls /app/configyamls
CMD ["api-server", "start"]
