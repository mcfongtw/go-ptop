FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y --no-install-recommends openjdk-17-jdk golang-go procps && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Cache go modules
COPY go.mod go.sum ./
RUN cat go.mod && go mod download

# Copy entire repo contents
COPY . .

# Build sample Java jar
RUN javac build/HelloWorld.java && \
    jar cfe /app/helloworld.jar HelloWorld -C build HelloWorld.class

# Build go-ptop binary
RUN go build -o /app/go-ptop

CMD ["bash", "-c", "set -euo pipefail; java -jar /app/helloworld.jar >/tmp/java.log 2>&1 & pid=$!; echo 'Started HelloWorld JVM PID:' $pid; sleep 3; /app/go-ptop $pid; kill $pid"]
