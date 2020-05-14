FROM golang:1.14
WORKDIR /app
ADD . .
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o serviceCommunicator
EXPOSE 11111
CMD ["./serviceCommunicator"]