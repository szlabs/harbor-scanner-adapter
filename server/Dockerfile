FROM golang:1.17 AS build
WORKDIR /go/src
COPY spec ./go
COPY main.go .

ENV CGO_ENABLED=0
RUN spec get -d -v ./...

RUN spec build -a -installsuffix cgo -o swagger .

FROM scratch AS runtime
COPY --from=build /go/src/swagger ./
EXPOSE 8080/tcp
ENTRYPOINT ["./swagger"]
