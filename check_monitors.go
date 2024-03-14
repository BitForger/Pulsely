package main

import (
	"database/sql"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type Monitor struct {
	Description       string
	UniqueId          string
	FailureThreshold  string
	DurationThreshold string
}

// StartMonitors Get all monitors
// Spawn goroutine to watch for new monitors
// Spawn goroutines for each monitor to check the heartbeat status
func StartMonitors() {
	log.Debug().Msg("Starting monitors")
	monitors := make(chan Monitor)

	isFirstRun := true

	// A map of active monitors
	// key: uniqueId
	activeMonitors := make(map[string]bool)

	ok, db, err := NewDbConnection()
	if !ok {
		log.Fatal().Err(err).Msg("unable to connect to database")
	}

	// spawn a goroutine to watch for new monitors
	go func() {
		// run a function on a schedule to check for new monitors
		// if a new monitor is found, add it to the monitors channel
		// if a monitor is removed, remove it from the monitors channel
		// if a monitor is updated, update the monitor in the monitors channel

		duration := 5 * time.Second
		timer := time.NewTimer(duration)
		for {
			select {
			case <-timer.C:
				var query = `SELECT description, uniqueId, failureThreshold, durationThreshold FROM monitors`

				// if this is not the first run, only get monitors that are not in the activeMonitors map
				if !isFirstRun && len(activeMonitors) > 0 {
					query += ` WHERE uniqueId NOT IN `
					query += `( `
					i := 0
					for range activeMonitors {
						query += `?`
						if i > 0 {
							query += `,`
						}
						i++
					}
					query += ` )`
				}

				stmt, prepareErr := db.Prepare(query)
				if prepareErr != nil {
					log.Error().Err(prepareErr).Msg("failed to prepare monitor watch query")
				}

				log.Debug().Str("query", query).Msg("running query")

				var rows *sql.Rows
				if !isFirstRun && len(activeMonitors) > 0 {
					ids := make([]any, 0, len(activeMonitors))
					for id := range activeMonitors {
						ids = append(ids, id)
					}
					log.Debug().Any("ids", ids).Msg("ids")

					var queryErr error
					rows, queryErr = stmt.Query(ids...)
					if queryErr != nil {
						log.Error().Err(queryErr).Msg("failed to execute query with ids")
					}
				} else {
					var queryErr error
					rows, queryErr = stmt.Query()
					if queryErr != nil {
						log.Error().Err(queryErr).Msg("query failed")
					}
				}

				defer rows.Close()
				var description string
				var uniqueId string
				var failureThreshold string
				var durationThreshold string
				for rows.Next() {
					if scanErr := rows.Scan(&description, &uniqueId, &failureThreshold, &durationThreshold); scanErr != nil {
						log.Error().Err(scanErr).Msg("failed to get row")
					}

					obj := Monitor{
						UniqueId:          uniqueId,
						Description:       description,
						FailureThreshold:  failureThreshold,
						DurationThreshold: durationThreshold,
					}
					activeMonitors[uniqueId] = true

					log.Debug().Str("monitor", obj.UniqueId).Msg("sending monitor to channel")
					monitors <- obj
				}

				isFirstRun = false
				timer.Reset(duration)
			}
		}
	}()

	var wg sync.WaitGroup
	for mon := range monitors {
		log.Debug().Str("monitor", mon.UniqueId).Msg("adding monitor")
		wg.Add(1)
		go func(m Monitor) {
			defer wg.Done()
			log.Debug().Str("monitor", m.UniqueId).Msg("starting monitor")
			// query the heartbeats table for all heartbeats within the duration threshold
			// if the number of heartbeats is greater than the failure threshold, mark the monitor as down
			const query = `SELECT COUNT(*) FROM heartbeats WHERE hookId = ? AND timestamp > ? AND up = 0`

			stmt, err := db.Prepare(query)
			if err != nil {
				log.Error().Err(err).Msg("failed to prepare heartbeat query")
			}

			stmt.Query(m.UniqueId, m.DurationThreshold)

		}(mon)
	}
}
