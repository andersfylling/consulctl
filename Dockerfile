FROM golang:1.11.2 as builder

WORKDIR $GOPATH/src/github.com/andersfylling/consulctl
COPY . .

RUN go get -d -v
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consulctl .
RUN ./consulctl --version
RUN mv consulctl /

FROM alpine:3.8
WORKDIR /
COPY --from=builder /consulctl /usr/bin
RUN consulctl --version