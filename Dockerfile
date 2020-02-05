FROM golang:onbuild AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM scratch
WORKDIR /root/
COPY --from=builder /go/src/app/app .
CMD ["./app"]
