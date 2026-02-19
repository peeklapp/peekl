FROM golang:1.26.0-trixie

WORKDIR /peekl

COPY . /peekl

CMD ["go", "test", "-coverprofile=coverage.out", "./..."]
