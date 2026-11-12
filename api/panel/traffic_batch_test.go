package panel

import (
	"context"
	"github.com/Sakawat-hossain/V2bX/conf"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatchAcknowledgementAndLegacyCompatibility(t *testing.T) {
	for _, ack := range []string{"", "0123456789abcdef0123456789abcdef", "different"} {
		t.Run("ack="+ack, func(t *testing.T) {
			id := "0123456789abcdef0123456789abcdef"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Traffic-Report-ID") != id {
					t.Error("missing report id")
				}
				if ack != "" {
					w.Header().Set("X-Traffic-Report-Ack", ack)
				}
				w.Write([]byte(`{"data":true}`))
			}))
			defer server.Close()
			c, err := New(&conf.ApiConfig{APIHost: server.URL, NodeType: "vless", NodeID: 1})
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			err = c.ReportTrafficBatchCtx(context.Background(), id, []UserTraffic{{UID: 1, Upload: 10}})
			if (err != nil) != (ack == "different") {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}
