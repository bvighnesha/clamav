FROM golang:1.13-onbuild
ENTRYPOINT ["/usr/local/bin/go-wrapper", "run"]
EXPOSE 8000
