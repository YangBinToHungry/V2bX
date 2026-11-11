//go:build hysteria2

package hy2

import (
	"github.com/Sakawat-hossain/V2bX/api/panel"
	"github.com/Sakawat-hossain/V2bX/conf"
	"github.com/Sakawat-hossain/V2bX/limiter"
	"go.uber.org/zap"
	"testing"
)

func TestEstablishedSessionStopsLoggingTrafficAfterRevocation(t *testing.T) {
	const tag = "test-revoked-session"
	user := panel.UserInfo{Id: 1, Uuid: "test-user"}
	l := limiter.AddLimiter("hysteria2", tag, &conf.LimitConfig{}, []panel.UserInfo{user}, map[int]int{})
	defer limiter.DeleteLimiter(tag)
	hook := &HookServer{Tag: tag, logger: zap.NewNop()}
	if !hook.LogTraffic(user.Uuid, 100, 100) {
		t.Fatal("authorized traffic must be accepted")
	}
	l.UpdateUser(tag, nil, []panel.UserInfo{user}, nil)
	if hook.LogTraffic(user.Uuid, 100, 100) {
		t.Fatal("revoked established session must return false so Hysteria disconnects it")
	}
}
