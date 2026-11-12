package node

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/Sakawat-hossain/V2bX/api/panel"
	log "github.com/sirupsen/logrus"
)

func (c *Controller) reportUserTrafficTask(ctx context.Context) (err error) {
	c.reportMu.Lock()
	defer c.reportMu.Unlock()
	userTraffic, fetchErr := c.server.GetUserTrafficSlice(c.tag, true)
	if fetchErr != nil {
		log.WithField("tag", c.tag).WithError(fetchErr).Warn("Read traffic counters failed; retry next cycle")
	}
	c.recordDynamicTraffic(userTraffic)
	c.queueTraffic(userTraffic)
	if c.pendingTraffic == nil && len(c.queuedTraffic) > 0 {
		id, idErr := newTrafficBatchID()
		if idErr != nil {
			log.WithField("tag", c.tag).WithError(idErr).Warn("Create traffic batch id failed; queued counters retained")
			return nil
		}
		n := min(1000, len(c.queuedTraffic))
		c.pendingTraffic = &trafficBatch{ID: id, Users: append([]panel.UserTraffic(nil), c.queuedTraffic[:n]...)}
		c.queuedTraffic = c.queuedTraffic[n:]
	}

	if batch := c.pendingTraffic; batch != nil {
		userTraffic = batch.Users
		if reportErr := c.getAPIClient().ReportTrafficBatchCtx(ctx, batch.ID, batch.Users); reportErr != nil {
			log.WithField("tag", c.tag).WithError(reportErr).Warn("Traffic batch retained for retry")
		} else {
			c.pendingTraffic = nil
			log.WithField("tag", c.tag).Infof("Report %d users traffic", len(batch.Users))
		}
	}

	if onlineDevice, err := c.limiter.GetOnlineDevice(); err != nil {
		log.Print(err)
	} else if len(onlineDevice) > 0 {
		var result []panel.OnlineUser
		var nocountUID = make(map[int]struct{})
		minTraffic := c.Options.DeviceOnlineMinTraffic
		// W2.4: atomic snapshot of NodeInfo; tolerate the brief window where
		// info is nil during startup.
		curInfo := c.info.Load()
		if curInfo != nil && curInfo.DeviceOnlineMinTraffic > 0 {
			minTraffic = int64(curInfo.DeviceOnlineMinTraffic)
		}
		for _, traffic := range userTraffic {
			total := traffic.Upload + traffic.Download
			if total < minTraffic*1000 {
				nocountUID[traffic.UID] = struct{}{}
			}
		}
		for _, online := range onlineDevice {
			if _, ok := nocountUID[online.UID]; !ok {
				result = append(result, online)
			}
		}
		data := make(map[int][]string)
		for _, onlineuser := range result {
			data[onlineuser.UID] = append(data[onlineuser.UID], onlineuser.IP)
		}
		if err = c.getAPIClient().ReportNodeOnlineUsersCtx(ctx, &data); err != nil {
			log.WithFields(log.Fields{
				"tag": c.tag,
				"err": err,
			}).Info("Report online users failed")
		} else {
			log.WithField("tag", c.tag).Infof("Total %d online users, %d Reported", len(onlineDevice), len(result))
			log.WithField("tag", c.tag).Debugf("Online users: %+v", data)
		}
	}

	userTraffic = nil
	return nil
}

func compareUserList(old, new []panel.UserInfo) (deleted, added, modified []panel.UserInfo) {
	oldMap := make(map[string]panel.UserInfo, len(old))
	for _, u := range old {
		oldMap[u.Uuid] = u
	}

	for _, u := range new {
		if o, ok := oldMap[u.Uuid]; !ok {
			added = append(added, u)
		} else {
			if o.SpeedLimit != u.SpeedLimit || o.DeviceLimit != u.DeviceLimit {
				modified = append(modified, u)
			}
			delete(oldMap, u.Uuid)
		}
	}

	for _, o := range oldMap {
		deleted = append(deleted, o)
	}

	return deleted, added, modified
}

// A batch is never merged with fresh counters while its result is uncertain.
// Pending batches live in memory; a process crash can still lose unreported traffic.
type trafficBatch struct {
	ID    string
	Users []panel.UserTraffic
}

func newTrafficBatchID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", fmt.Errorf("create traffic batch id: %w", err)
	}
	return hex.EncodeToString(id[:]), nil
}

func (c *Controller) dynamicSpeedLimitEnabled() bool {
	d := c.LimitConfig.DynamicSpeedLimitConfig
	return c.LimitConfig.EnableDynamicSpeedLimit && d != nil && d.Periodic > 0 && d.Traffic > 0 && d.SpeedLimit > 0 && d.ExpireTime > 0
}

// Traffic is measured in bytes and counted once, when counters are extracted,
// not on each retry of an unacknowledged report.
func (c *Controller) recordDynamicTraffic(traffic []panel.UserTraffic) {
	if !c.dynamicSpeedLimitEnabled() || c.limiter == nil {
		return
	}
	totals := make(map[int]int64, len(traffic))
	for _, t := range traffic {
		totals[t.UID] += t.Upload + t.Download
	}
	c.trafficMu.Lock()
	defer c.trafficMu.Unlock()
	if c.traffic == nil {
		c.traffic = make(map[string]int64)
	}
	c.limiter.UUIDtoUID.Range(func(key, value any) bool {
		if amount := totals[value.(int)]; amount > 0 {
			c.traffic[key.(string)] += amount
		}
		return true
	})
}

// Aggregate fresh deltas separately while the immutable in-flight batch retries.
// Memory grows with distinct users, rather than the number of failed cycles.
func (c *Controller) queueTraffic(fresh []panel.UserTraffic) {
	if len(fresh) == 0 {
		return
	}
	totals := make(map[int]panel.UserTraffic, len(c.queuedTraffic)+len(fresh))
	for _, list := range [][]panel.UserTraffic{c.queuedTraffic, fresh} {
		for _, t := range list {
			v := totals[t.UID]
			v.UID = t.UID
			v.Upload += t.Upload
			v.Download += t.Download
			totals[t.UID] = v
		}
	}
	c.queuedTraffic = make([]panel.UserTraffic, 0, len(totals))
	for _, t := range totals {
		c.queuedTraffic = append(c.queuedTraffic, t)
	}
}
