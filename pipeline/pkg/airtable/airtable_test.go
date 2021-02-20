package airtable

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestObjectFromFile(t *testing.T) {
	tests := []struct {
		desc       string
		filename   string
		wantLength int
		wantErr    bool
	}{
		{
			desc:       "success",
			filename:   "test_data/two_counties.json",
			wantLength: 2,
		},
		{
			desc:     "file does not exist",
			filename: "test_data/doesnotexist.json",
			wantErr:  true,
		},
		{
			desc:     "not valid json",
			filename: "test_data/notvalidjson.json",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx := context.Background()
			o, err := ObjectFromFile(ctx, "counties", tt.filename)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
			if len(o) != tt.wantLength {
				t.Errorf("got %v records, want %v", len(o), tt.wantLength)
			}

		})
	}
}

type stubHTTP struct {
	seq   int
	resps []*http.Response
}

func (s *stubHTTP) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := s.resps[s.seq]
	s.seq++
	return resp, nil
}

type stubFailHTTP struct{}

func (s *stubFailHTTP) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("round trip failure")
}

func TestDownload(t *testing.T) {
	ctx := context.Background()

	raw, err := ioutil.ReadFile("test_data/counties.raw")
	if err != nil {
		t.Fatalf("can't read test data counties.raw: %v", err)
	}

	tests := []struct {
		desc    string
		transp  http.RoundTripper
		wantErr bool
		wantLen int
		table   string
	}{
		{
			desc: "one request, success",
			transp: &stubHTTP{
				resps: []*http.Response{
					&http.Response{
						StatusCode:    http.StatusOK,
						Body:          ioutil.NopCloser(bytes.NewBuffer(raw)),
						ContentLength: int64(len(raw)),
					},
				},
			},
			wantLen: 3,
		},
		{
			desc: "500 error",
			transp: &stubHTTP{
				resps: []*http.Response{
					&http.Response{
						StatusCode:    http.StatusInternalServerError,
						Body:          ioutil.NopCloser(bytes.NewBuffer([]byte{})),
						ContentLength: 0,
					},
				},
			},
			wantErr: true,
		},
		{
			desc:    "invalid URL path",
			table:   "%percent-is-for-encoding%",
			wantErr: true,
		},
		{
			desc: "backoff, then succeed",
			transp: &stubHTTP{
				resps: []*http.Response{
					&http.Response{
						StatusCode:    http.StatusTooManyRequests,
						Body:          ioutil.NopCloser(bytes.NewBuffer([]byte{})),
						ContentLength: 0,
					},
					&http.Response{
						StatusCode:    http.StatusOK,
						Body:          ioutil.NopCloser(bytes.NewBuffer(raw)),
						ContentLength: int64(len(raw)),
					},
				},
			},
		},
		{
			desc: "invalid json",
			transp: &stubHTTP{
				resps: []*http.Response{
					&http.Response{
						StatusCode:    http.StatusOK,
						Body:          ioutil.NopCloser(bytes.NewBufferString("hisdfsf")),
						ContentLength: int64(len(raw)),
					},
				},
			},
			wantErr: true,
		},
		{
			desc:    "other http request failure",
			transp:  &stubFailHTTP{},
			wantErr: true,
		},
		{
			desc: "multiple requests",
			transp: &stubHTTP{
				resps: []*http.Response{
					&http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(
							`{"offset":"1", "records":[{"id":"recA","fields":{"County":"Glenn County"}}]}`)),
					},
					&http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(
							`{"records":[{"id":"recB","fields":{"County":"Another County"}}]}`)),
					},
				},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tables := &airtable{
				httpClient: &http.Client{Transport: tt.transp},
			}

			tn := tt.table
			if tn == "" {
				tn = "counties"
			}

			content, err := tables.Download(ctx, tn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Download() unexpected error: %v", err)
			}
			if len(content) != tt.wantLen {
				t.Errorf("got %v counties, want %v", len(content), tt.wantLen)
			}
		})
	}

}
