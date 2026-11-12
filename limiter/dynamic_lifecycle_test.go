package limiter

import (
	"github.com/Sakawat-hossain/V2bX/api/panel"
	"github.com/Sakawat-hossain/V2bX/common/format"
	"github.com/Sakawat-hossain/V2bX/conf"
	"testing"
	"time"
)

func TestDynamicLimitExpiryRetainsAuthorizationAndSharedBucket(t *testing.T) {
	tag := t.Name()
	user := panel.UserInfo{Id: 1, Uuid: "user"}
	l := AddLimiter("vless", tag, &conf.LimitConfig{EnableDynamicSpeedLimit: true}, []panel.UserInfo{user}, nil)
	defer DeleteLimiter(tag)
	key := format.UserTag(tag, user.Uuid)
	bucket, rejected := l.CheckLimit(key, "127.0.0.1", true, false)
	if rejected || bucket == nil || bucket.Get() != nil {
		t.Fatal("unlimited connections must keep a hot-swappable handle")
	}
	if err := l.UpdateDynamicSpeedLimit(tag, user.Uuid, 5, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if bucket.Get() == nil {
		t.Fatal("existing connection did not receive the dynamic limit")
	}
	limited := user
	limited.SpeedLimit = 2
	l.UpdateUser(tag, nil, nil, []panel.UserInfo{limited})
	if bucket.Get().Capacity() != 250000 {
		t.Fatal("static changes must retain the stricter active limit")
	}
	l.UpdateUser(tag, nil, nil, []panel.UserInfo{user})
	if bucket.Get() == nil || bucket.Get().Capacity() != 625000 {
		t.Fatal("removing a static limit must not remove the active dynamic limit")
	}
	v, _ := l.UserLimitInfo.Load(key)
	v.(*UserLimitInfo).ExpireTime.Store(time.Now().Add(-time.Second).Unix())
	l.ExpireDynamicSpeedLimits(time.Now())
	if bucket.Get() != nil {
		t.Fatal("periodic expiry must restore a long-lived connection without a new handshake")
	}
	got, rejected := l.CheckLimit(key, "127.0.0.1", true, false)
	if rejected || got != bucket || bucket.Get() != nil {
		t.Fatal("expiry must restore unlimited speed without replacing the connection handle")
	}
	if _, ok := l.UserLimitInfo.Load(key); !ok {
		t.Fatal("expiry incorrectly revoked the user")
	}
}
