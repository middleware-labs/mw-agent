FROM busybox:glibc 
RUN mkdir -p /mwapp
WORKDIR /mwapp
COPY . /mwapp
# ENTRYPOINT ["/mwapp"]
