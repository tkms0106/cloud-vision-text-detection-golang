FROM golang:1.15.0 as dev
RUN mkdir -p /go/src/github.com/tkms0106/
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
WORKDIR /go/src/github.com/tkms0106/cloud-vision-text-detection-golang
RUN apt update \
 && apt install -y curl \
 && curl -fLo /bin/air https://github.com/cosmtrek/air/releases/download/v1.12.3/air_1.12.3_linux_amd64 \
 && chmod +x /bin/air \
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
