FROM golang:1.27.0-trixie

WORKDIR /peekl

COPY . /peekl

CMD ["go", "test", "-coverprofile=coverage.out", "./..."]
