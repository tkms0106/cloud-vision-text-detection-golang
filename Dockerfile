FROM golang:1.15.6 as dev
RUN mkdir -p /go/src/github.com/tkms0106/
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
WORKDIR /go/src/github.com/tkms0106/cloud-vision-text-detection-golang
RUN apt update \
 && apt install -y curl \
 && curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.12.4 \
 && apt remove -y curl
COPY go.mod go.sum ./
RUN go mod download

FROM dev as builder
RUN groupadd -g 1000 app \
 && useradd -u 1000 -g app app
COPY ./main.go ./main.go
RUN go build ./main.go

FROM gcr.io/distroless/base-debian10 as runner
WORKDIR /home/tkms0106/app
COPY ./assets ./assets
COPY --from=builder /go/src/github.com/tkms0106/cloud-vision-text-detection-golang/main .
USER app
ENTRYPOINT ["./main"]
