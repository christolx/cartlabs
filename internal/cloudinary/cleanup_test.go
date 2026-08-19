package cloudinary

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteDemoUploadsFollowsCursor(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodDelete || r.URL.Path != "/v1_1/demo/resources/image/tags/cartlabs_demo_upload" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte("key:secret")) {
			t.Fatal("missing basic authentication")
		}
		w.Header().Set("Content-Type", "application/json")
		if requests == 1 {
			_, _ = fmt.Fprint(w, `{"deleted":{"first":"deleted"},"partial":true,"next_cursor":"page-2"}`)
			return
		}
		if r.URL.Query().Get("next_cursor") != "page-2" {
			t.Fatalf("cursor = %q", r.URL.Query().Get("next_cursor"))
		}
		_, _ = fmt.Fprint(w, `{"deleted":{"second":"deleted"},"partial":false}`)
	}))
	defer server.Close()

	result, err := DeleteDemoUploads(context.Background(), server.Client(), CleanupConfig{CloudName: "demo", APIKey: "key", APISecret: "secret", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if result.Deleted != 2 || result.Pages != 2 || requests != 2 {
		t.Fatalf("result=%#v requests=%d", result, requests)
	}
}
