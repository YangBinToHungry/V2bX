package panel

import (
	"github.com/go-resty/resty/v2"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserListEmptyVersusUnchanged(t *testing.T) {
	for _, format := range []string{"json", "msgpack"} {
		t.Run(format, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("If-None-Match") == "empty-users" {
					w.WriteHeader(http.StatusNotModified)
					return
				}
				w.Header().Set("ETag", "empty-users")
				if format == "msgpack" {
					w.Header().Set("Content-Type", "application/x-msgpack")
					body, err := msgpack.Marshal(UserListBody{Users: []UserInfo{}})
					if err != nil {
						t.Fatal(err)
					}
					w.Write(body)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"users":[]}`))
				}
			}))
			defer server.Close()
			c := &Client{client: resty.New().SetBaseURL(server.URL)}
			users, err := c.GetUserList()
			if err != nil {
				t.Fatal(err)
			}
			if users == nil || len(users) != 0 {
				t.Fatalf("empty HTTP 200 must be non-nil, got %#v", users)
			}
			users, err = c.GetUserList()
			if err != nil {
				t.Fatal(err)
			}
			if users != nil {
				t.Fatalf("HTTP 304 must be nil, got %#v", users)
			}
		})
	}
}
