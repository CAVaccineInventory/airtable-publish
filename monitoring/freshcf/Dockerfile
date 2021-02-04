FROM alpine:latest

RUN apk add --no-cache go

# Cache the download of dependent modules.
# Only copy the go.* files so source code changes don't result in new downloads.
COPY ./go.* /src/
WORKDIR /src
RUN go mod download

# Copy the rest of the source code into the container.
COPY ./ /src

# Build!
RUN go build -o /freshcf ./monitoring/freshcf/cmd/server/main.go

# Setup runtime environment
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

CMD ["/entrypoint.sh", "freshcf"]
EXPOSE 8080
