package conf

import (
	"testing"
)

func TestConf_LoadFromPath(t *testing.T) {
	c := New()
	t.Log(c.LoadFromPath("../example/config.json"), c.NodeConfig)
}

func TestConf_Watch(t *testing.T) {
	if testing.Short() {
		t.Skip("manual watcher test: blocks forever (select{}) to observe live reloads")
	}
	c := New()
	t.Log(c.Watch("./1.json", "", "", func() {}))
	select {}
}
