FROM golang:1.14 as builder
WORKDIR /go/src/app
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./http.go ./
COPY ./main.go ./
COPY ./config.go ./
COPY ./stream.go ./
RUN go get
RUN go install

FROM ubuntu:focal
WORKDIR /app
COPY ./config.json ./
COPY ./web ./web
COPY ./doc ./doc
COPY --from=builder /go/bin/RTSPtoWebRTC ./
CMD /app/RTSPtoWebRTC