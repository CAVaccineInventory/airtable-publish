FROM gcr.io/google.com/cloudsdktool/cloud-sdk:alpine

RUN apk add --no-cache py3-pip go jq && \
	pip3 install airtable-export

COPY *.go go.mod go.sum /
RUN go build

COPY entrypoint.sh /
RUN chmod +x entrypoint.sh

CMD ["sh", "entrypoint.sh"]
