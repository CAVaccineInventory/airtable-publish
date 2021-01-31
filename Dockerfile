FROM gcr.io/google.com/cloudsdktool/cloud-sdk:alpine

RUN apk add --no-cache py3-pip go jq && \
	pip3 install airtable-export

COPY pipeline/go.* /
RUN go mod download

COPY pipeline/ /
RUN go build -o server ./cmd/server/main.go
RUN go build -o once   ./cmd/once/main.go

COPY entrypoint.sh /
RUN chmod +x entrypoint.sh

CMD ["sh", "entrypoint.sh"]
EXPOSE 8080
