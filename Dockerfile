FROM golang:alpine

RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN apk add make

RUN make build

CMD ["./vedran"]