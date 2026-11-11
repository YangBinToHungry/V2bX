package node

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Sakawat-hossain/V2bX/api/panel"
	"github.com/Sakawat-hossain/V2bX/common/format"
	"github.com/Sakawat-hossain/V2bX/conf"
	vCore "github.com/Sakawat-hossain/V2bX/core"
	"github.com/Sakawat-hossain/V2bX/limiter"
)

type reliabilityCore struct {
	vCore.Core
	failAdd   bool
	deleted   int
	snapshots [][]panel.UserTraffic
	fetches   int
}

func (f *reliabilityCore) DelUsers(users []panel.UserInfo, _ string, _ *panel.NodeInfo) error {
	f.deleted += len(users)
	return nil
}
func (f *reliabilityCore) AddUsers(p *vCore.AddUsersParams) (int, error) {
	if f.failAdd {
		return 0, errors.New("injected add failure")
	}
	return len(p.Users), nil
}
func (f *reliabilityCore) GetUserTrafficSlice(_ string, _ bool) ([]panel.UserTraffic, error) {
	i := f.fetches
	f.fetches++
	if i >= len(f.snapshots) {
		return nil, nil
	}
	return f.snapshots[i], nil
}
func newReliabilityController(t *testing.T, url string, core *reliabilityCore, users []panel.UserInfo) *Controller {
	t.Helper()
	api, err := panel.New(&conf.ApiConfig{APIHost: url, NodeType: "vless", NodeID: 1, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(api.Close)
	tag := t.Name()
	l := limiter.AddLimiter("vless", tag, &conf.LimitConfig{}, users, map[int]int{})
	t.Cleanup(func() { limiter.DeleteLimiter(tag) })
	c := &Controller{apiClient: api, server: core, tag: tag, limiter: l, userList: append([]panel.UserInfo(nil), users...), Options: &conf.Options{}}
	c.info.Store(&panel.NodeInfo{Type: "vless"})
	return c
}

func TestUsersRefetchedAfterAliveFailure(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	staleHeader := false
	aliveFailure := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/api/v1/server/UniProxy/config":
			w.WriteHeader(http.StatusNotModified)
		case "/api/v1/server/UniProxy/user":
			calls++
			if calls == 2 && r.Header.Get("If-None-Match") != "" {
				staleHeader = true
			}
			if r.Header.Get("If-None-Match") == "changed" {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("ETag", "changed")
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"users":[]}`)
		case "/api/v1/server/UniProxy/alivelist":
			if aliveFailure {
				conn, _, _ := w.(http.Hijacker).Hijack()
				conn.Close()
				return
			}
			io.WriteString(w, `{"alive":{}}`)
		}
	}))
	defer server.Close()
	core := &reliabilityCore{}
	c := newReliabilityController(t, server.URL, core, []panel.UserInfo{{Id: 1, Uuid: "revoked"}})
	c.nodeInfoMonitor(context.Background())
	mu.Lock()
	aliveFailure = false
	mu.Unlock()
	if core.deleted != 0 {
		t.Fatal("must not apply a failed cycle")
	}
	c.nodeInfoMonitor(context.Background())
	if staleHeader || core.deleted != 1 || len(c.userList) != 0 {
		t.Fatalf("update was lost: stale=%v deleted=%d", staleHeader, core.deleted)
	}
}

func TestPartialUserApplyDoesNotRepeatSuccessfulDeletion(t *testing.T) {
	core := &reliabilityCore{failAdd: true}
	c := newReliabilityController(t, "http://127.0.0.1", core, []panel.UserInfo{{Id: 1, Uuid: "old"}})
	desired := []panel.UserInfo{{Id: 2, Uuid: "new"}}
	if c.applyUserList(desired) == nil {
		t.Fatal("expected injected failure")
	}
	if len(c.userList) != 0 {
		t.Fatal("successful deletion must be committed locally")
	}
	core.failAdd = false
	if err := c.applyUserList(desired); err != nil {
		t.Fatal(err)
	}
	if core.deleted != 1 || len(c.userList) != 1 {
		t.Fatal("retry repeated an already applied operation")
	}
}

func TestLostTrafficAcknowledgementRetainsImmutableBatch(t *testing.T) {
	var ids, bodies []string
	seen := map[string]bool{}
	charged := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/server/UniProxy/push" {
			io.WriteString(w, `{}`)
			return
		}
		b, _ := io.ReadAll(r.Body)
		id := r.Header.Get("X-Traffic-Report-ID")
		ids = append(ids, id)
		bodies = append(bodies, string(b))
		if !seen[id] {
			seen[id] = true
			charged++
		}
		if len(ids) == 1 {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("X-Traffic-Report-Ack", id)
		io.WriteString(w, `{"data":true}`)
	}))
	defer server.Close()
	core := &reliabilityCore{snapshots: [][]panel.UserTraffic{{{UID: 1, Upload: 10}}, {{UID: 1, Upload: 20}}}}
	c := newReliabilityController(t, server.URL, core, []panel.UserInfo{{Id: 1, Uuid: "user"}})
	c.LimitConfig = conf.LimitConfig{EnableDynamicSpeedLimit: true, DynamicSpeedLimitConfig: &conf.DynamicSpeedLimitConfig{Periodic: 1, Traffic: 100, SpeedLimit: 5, ExpireTime: 1}}
	for i := 0; i < 2; i++ {
		if err := c.reportUserTrafficTask(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if ids[0] == "" || ids[0] != ids[1] || bodies[0] != bodies[1] || core.fetches != 2 || charged != 1 || c.traffic["user"] != 30 {
		t.Fatal("retry changed batch identity, merged fresh counters into pending batch, or counted bytes twice")
	}
	c.reportUserTrafficTask(context.Background())
	if ids[2] == ids[0] || charged != 2 || c.traffic["user"] != 30 {
		t.Fatal("fresh counters must use a fresh report id")
	}
}

func TestDynamicLimitTaskStartsAndConsumesTraffic(t *testing.T) {
	c := newReliabilityController(t, "http://127.0.0.1", &reliabilityCore{}, []panel.UserInfo{{Id: 1, Uuid: "user"}})
	c.LimitConfig = conf.LimitConfig{EnableDynamicSpeedLimit: true, DynamicSpeedLimitConfig: &conf.DynamicSpeedLimitConfig{Periodic: 1, Traffic: 10, SpeedLimit: 5, ExpireTime: 1}}
	c.startTasks(&panel.NodeInfo{PullInterval: time.Hour, PushInterval: time.Hour})
	defer c.nodeInfoMonitorPeriodic.Close()
	defer c.userReportPeriodic.Close()
	defer c.dynamicSpeedLimitPeriodic.Close()
	c.recordDynamicTraffic([]panel.UserTraffic{{UID: 1, Upload: 11}})
	v, _ := c.limiter.UserLimitInfo.Load(format.UserTag(c.tag, "user"))
	limit := v.(*limiter.UserLimitInfo)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if limit.DynamicSpeedLimit.Load() == 5 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("dynamic limiter was never run")
}
