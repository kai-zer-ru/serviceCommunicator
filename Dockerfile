FROM golang:latest
WORKDIR /app
ADD . .
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o serviceCommunicator
EXPOSE 4000
CMD ["./serviceCommunicator"]