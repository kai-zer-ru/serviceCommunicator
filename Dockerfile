FROM golang:1.14
WORKDIR /app
ADD . .
EXPOSE 4000
CMD ["./serviceCommunicator"]