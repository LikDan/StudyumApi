FROM golang:1.19-alpine as builder

WORKDIR /app

COPY .. ./
COPY ../cmd/studyum ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server

FROM scratch

COPY ../email-templates ./email-templates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server

CMD ["/server"]