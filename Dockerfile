FROM busybox:glibc 
RUN mkdir -p /mwapp
WORKDIR /mwapp
COPY . /mwapp
CMD ["./agent-host-go start"]
