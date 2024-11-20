FROM golang:1.23.3-alpine3.20

ENV SOURCES=/go/src/github.com/blueorb/microservice

COPY . ${SOURCES}

RUN cd ${SOURCES} && go build -o /bin/microservice

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT microservice
