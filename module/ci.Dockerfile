FROM ubuntu:latest

# Install ca-certificates
RUN apt get install ca-certificates

EXPOSE 26656 26657 1317 9090

COPY build/mhub2 /usr/bin/mhub2

CMD ["mhub2", "start"]