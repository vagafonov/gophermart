FROM golang:1.21
WORKDIR /opt/app/api
ENV GOPATH=/
COPY ./ ./
RUN pwd
RUN ls
ENTRYPOINT ["go", "run", "./cmd/gophermart/main.go"]