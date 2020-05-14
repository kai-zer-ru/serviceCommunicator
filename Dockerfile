FROM golang:1.14
WORKDIR /app
ADD . .

EXPOSE 11111
CMD ["./serviceCommunicator"]