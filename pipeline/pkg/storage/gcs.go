package storage

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints/metadata"
	beeline "github.com/honeycombio/beeline-go"
)

// Update the README.md for new latencies if you adjust the max-age.
const cacheControl = "public,max-age=120"

// UploadToGCS uploads to GCS, after gzip'ing and setting a cache-control header.
func UploadToGCS(ctx context.Context, destinationFile string, transformedData metadata.JSONData) error {
	ctx, span := beeline.StartSpan(ctx, "storage.UploadToGCS")
	defer span.Send()
	beeline.AddField(ctx, "destinationFile", destinationFile)

	serializedData, err := Serialize(transformedData)
	if err != nil {
		err = fmt.Errorf("failed to write serialized json: %w", err)
		beeline.AddField(ctx, "error", err)
		return err
	}

	gb := &bytes.Buffer{}
	gw := gzip.NewWriter(gb)
	_, err = gw.Write(serializedData.Bytes())
	if err != nil {
		err = fmt.Errorf("failed to compress serialized json: %w", err)
		beeline.AddField(ctx, "error", err)
		return err
	}
	gw.Close()
	if err != nil {
		err = fmt.Errorf("failed to close compressed serialized json: %w", err)
		beeline.AddField(ctx, "error", err)
		return err
	}

	bucket, object, err := parts(destinationFile)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return err
	}
	err = uploadFile(ctx, bucket, object, gb.Bytes(),
		WithContentEncoding("gzip"),
		WithCacheControl(cacheControl),
		WithContentType("application/json"))
	if err != nil {
		err = fmt.Errorf("failed to upload file: %w", err)
		beeline.AddField(ctx, "error", err)
		return err
	}

	return nil
}

// UploadOptionFunc is used to support named parameters to uploadFile.  It's exported to quiet a lint warning, but otherwise doesn't need to be because it's only used by an unexported function.
type UploadOptionFunc func(w *storage.Writer)

// WithCacheControl returns an uploadOptionFunc that sets the CacheControl field on the provided storage.Writer.
func WithCacheControl(cc string) UploadOptionFunc {
	return func(w *storage.Writer) { w.CacheControl = cc }
}

// WithContentEncoding returns an uploadOptionFunc that sets the ContentEncoding field on the provided storage.Writer.
func WithContentEncoding(ce string) UploadOptionFunc {
	return func(w *storage.Writer) { w.ContentEncoding = ce }
}

// WithContentType returns an uploadOptionFunc that sets the ContentType field on the provided storage.Writer.
func WithContentType(ct string) UploadOptionFunc {
	return func(w *storage.Writer) { w.ContentType = ct }
}

// parts splits a gs:// URI into component parts (bucket, object)
func parts(p string) (string, string, error) {
	if !strings.HasPrefix(p, "gs://") {
		return "", "", fmt.Errorf("missing gs prefix on %q", p)
	}

	ps := strings.SplitN(p, "/", 4)
	return ps[2], ps[3], nil
}

// uploadFile uploads an object to Cloud Storage
func uploadFile(ctx context.Context, bucket, object string, content []byte, opts ...UploadOptionFunc) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	br := bytes.NewReader(content)

	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	for _, f := range opts {
		f(wc)
	}

	if _, err = io.Copy(wc, br); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	return nil
}
