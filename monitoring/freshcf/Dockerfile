FROM gcr.io/google.com/cloudsdktool/cloud-sdk:alpine

RUN apk add --no-cache go 

# Cache the download of dependent modules.  
# Only copy the go.* files so source code changes don't result in new downloads.
COPY ./go.* /src/
WORKDIR /src
RUN go mod download

# Copy the source code into the container.
COPY ./pipeline/pkg/       /src/pipeline/pkg/
COPY ./monitoring/freshcf/ /src/monitoring/freshcf/

# Build!  $COMMIT_SHA is filled in by Google Build, or the `scripts/`
# build step.
ARG COMMIT_SHA
RUN go build \
        -ldflags "-X github.com/CAVaccineInventory/airtable-export/pipeline/pkg/config.GitCommit=$COMMIT_SHA" \
        -o /freshcf \
        ./monitoring/freshcf/cmd/server/main.go

# Setup runtime environment
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

CMD ["/entrypoint.sh", "freshcf"]
EXPOSE 8080
