FROM golang:latest
LABEL maintainer="Kaizer <kaizer@kai-zer.ru>"
WORKDIR /app
COPY go.mod go.sum .env ./
RUN go mod download
COPY . .
RUN go build -o serviceCommunicator .
EXPOSE 4000
CMD ["./serviceCommunicator"]