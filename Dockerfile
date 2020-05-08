FROM golang:latest
ADD .
WORKDIR /app
COPY go.mod go.sum .env ./
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o serviceCommunicator
EXPOSE 4000
CMD ["./serviceCommunicator"]