//go:build hysteria2

package hy2

import (
	"github.com/Sakawat-hossain/V2bX/api/panel"
	vCore "github.com/Sakawat-hossain/V2bX/core"
	"testing"
)

func TestNodeAuthenticationAndRevocationAreIsolated(t *testing.T) {
	h := &Hysteria2{Hy2nodes: map[string]Hysteria2node{
		"a": {Auth: &V2bX{usersMap: map[string]int{}}},
		"b": {Auth: &V2bX{usersMap: map[string]int{}}},
	}}
	shared := panel.UserInfo{Id: 1, Uuid: "shared"}
	onlyA := panel.UserInfo{Id: 2, Uuid: "only-a"}
	h.AddUsers(&vCore.AddUsersParams{Tag: "a", Users: []panel.UserInfo{shared, onlyA}})
	h.AddUsers(&vCore.AddUsersParams{Tag: "b", Users: []panel.UserInfo{shared}})
	if ok, _ := h.Hy2nodes["b"].Auth.Authenticate(nil, onlyA.Uuid, 0); ok {
		t.Fatal("node b accepted node a's user")
	}
	if err := h.DelUsers([]panel.UserInfo{shared}, "a", nil); err != nil {
		t.Fatal(err)
	}
	if ok, _ := h.Hy2nodes["a"].Auth.Authenticate(nil, shared.Uuid, 0); ok {
		t.Fatal("revoked node a user still authorized")
	}
	if ok, _ := h.Hy2nodes["b"].Auth.Authenticate(nil, shared.Uuid, 0); !ok {
		t.Fatal("revoking node a removed node b authorization")
	}
	if _, err := h.AddUsers(&vCore.AddUsersParams{Tag: "missing", Users: []panel.UserInfo{shared}}); err == nil {
		t.Fatal("unknown node must reject updates")
	}
}
