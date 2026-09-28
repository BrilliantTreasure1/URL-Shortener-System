package database

import (
	"database/sql"
	"log"
	"time"
)

func StartPoolStatsLogger(db *sql.DB, interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				stats := db.Stats()
				log.Printf("pool: Open=%d InUse=%d Idle=%d WaitCount=%d WaitDuration=%s MaxOpen=%d",
					stats.OpenConnections, stats.InUse, stats.Idle,
					stats.WaitCount, stats.WaitDuration, stats.MaxOpenConnections)
			case <-stop:
				return
			}
		}
	}()
}