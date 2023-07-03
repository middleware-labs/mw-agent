FROM golang:1.18 as base
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go get -d -v ./... && go mod tidy
RUN CGO_ENABLED=0 go build -o /tmp/api-server ./*.go

FROM busybox:glibc as prod
RUN mkdir -p /var/log
WORKDIR /app
COPY --from=base /tmp/api-server /usr/bin/api-server
COPY --from=base /app/configyamls /app/configyamls
CMD ["api-server", "start"]
