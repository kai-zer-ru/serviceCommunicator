FROM golang:1.14-alpine
WORKDIR /app
ADD . .
EXPOSE 4000
CMD ["./serviceCommunicator"]