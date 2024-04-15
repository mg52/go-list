FROM golang:1.22

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /golist-app

EXPOSE 8080

CMD ["/golist-app"]