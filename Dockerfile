FROM golang:alpine
#install mysqldump to docker
RUN apk add --no-cache mysql-client
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main ./cmd
#run dumper
CMD ["/app/main"]