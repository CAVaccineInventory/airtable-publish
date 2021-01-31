FROM gcr.io/google.com/cloudsdktool/cloud-sdk:alpine

RUN apk add --no-cache py3-pip go jq && \
	pip3 install airtable-export

# Cache the download of dependent modules.  
# Only copy the go.* files so source code changes don't result in new downloads.
COPY ./go.* /src/
WORKDIR /src
RUN go mod download

# Copy the rest of the source code into the container.
COPY ./ /src

# Build!
RUN go build -o /server ./pipeline/cmd/server/main.go
RUN go build -o /once   ./pipeline/cmd/once/main.go

# Setup runtime environment
COPY entrypoint.sh /
RUN chmod +x entrypoint.sh

CMD ["sh", "entrypoint.sh"]
EXPOSE 8080
