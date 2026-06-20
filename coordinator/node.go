package coordinator

import (
	"context"
	"log"
	"time"
)

// StartHeartbeatLoop periodically updates the node's last_heartbeat timestamp.
// Runs every 30 seconds until the context is cancelled or Stop() is called.
func (c *Coordinator) StartHeartbeatLoop(ctx context.Context) {
	if !c.IsPooled() || c.Client == nil {
		return
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stopCh:
				return
			case <-ticker.C:
				// Skip the heartbeat while draining/Stopping so we don't
				// clobber the draining status the shutdown path just set.
				c.mu.Lock()
				draining := c.draining
				c.mu.Unlock()
				if draining {
					continue
				}

				load := c.currentLoad()
				if err := c.Client.HeartbeatNode(c.NodeID, load); err != nil {
					log.Printf("[coordinator] heartbeat error: %v", err)
					continue
				}
				// Recover from a "stuck offline" state (e.g. the reaper marked
				// us offline during a restart gap). EnsureNodeOnline only
				// patches when status is not already online/draining, so it
				// never fights the draining flow.
				if err := c.Client.EnsureNodeOnline(c.NodeID); err != nil {
					log.Printf("[coordinator] ensure-online error: %v", err)
				}
			}
		}
	}()
}
