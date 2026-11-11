package node

import (
	"log"
	"os"
	"testing"

	"github.com/Sakawat-hossain/V2bX/conf"
)

var l *Lego

func init() {
	var err error
	l, err = NewLego(&conf.CertConfig{
		CertMode:   "dns",
		Email:      "test@test.com",
		CertDomain: "test.test.com",
		Provider:   "cloudflare",
		DNSEnv: map[string]string{
			"CF_DNS_API_TOKEN": "123",
		},
		CertFile: "./cert/1.pem",
		KeyFile:  "./cert/1.key",
	})
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func TestLego_CreateCertByDns(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs real ACME + DNS provider credentials")
	}
	err := l.CreateCert()
	if err != nil {
		t.Error(err)
	}
}

func TestLego_RenewCert(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs real ACME + DNS provider credentials")
	}
	log.Println(l.RenewCert())
}
