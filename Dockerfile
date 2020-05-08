FROM alpine:latest
WORKDIR /app
ADD . .
EXPOSE 4000
CMD ["./serviceCommunicator"]