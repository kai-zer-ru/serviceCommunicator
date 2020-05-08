FROM ubuntu:18.04
WORKDIR /app
ADD . .
EXPOSE 4000
CMD ["./serviceCommunicator"]