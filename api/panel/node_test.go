package panel

import (
	"log"
	"testing"

	"github.com/Sakawat-hossain/V2bX/conf"
)

var client *Client

func init() {
	c, err := New(&conf.ApiConfig{
		APIHost:  "http://127.0.0.1",
		Key:      "token",
		NodeType: "V2ray",
		NodeID:   1,
	})
	if err != nil {
		log.Panic(err)
	}
	client = c
}

func TestClient_GetNodeInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("integration smoke test: needs a panel reachable at 127.0.0.1")
	}
	log.Println(client.GetNodeInfo())
	log.Println(client.GetNodeInfo())
}

func TestClient_ReportUserTraffic(t *testing.T) {
	if testing.Short() {
		t.Skip("integration smoke test: needs a panel reachable at 127.0.0.1")
	}
	log.Println(client.ReportUserTraffic([]UserTraffic{
		{
			UID:      10372,
			Upload:   1000,
			Download: 1000,
		},
	}))
}
