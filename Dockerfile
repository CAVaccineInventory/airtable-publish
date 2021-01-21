FROM gcr.io/google.com/cloudsdktool/cloud-sdk:alpine

RUN apk add --no-cache py3-pip go jq && \
	pip3 install airtable-export

COPY ./sanitize ./sanitize
WORKDIR sanitize
RUN go build
WORKDIR ..

COPY sync.sh main.sh ./
RUN chmod +x sync.sh main.sh

CMD ["sh", "main.sh"]
