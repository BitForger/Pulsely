package main

import (
	"github.com/rs/zerolog/log"
)

type Monitor struct {
	Description string
	UniqueId    string
}

// Get all monitors
// Spawn goroutine to watch for new monitors
// Spawn goroutines for each monitor to check the heartbeat status
func start() {
	monitors := make(chan []Monitor)
	go func(mons []Monitor) {
		for _, mon := range mons {
			go func(m Monitor) {

			}(mon)
		}
	}(<-monitors)

	go func() {
		const query = `SELECT description, uniqueId FROM monitors`
		ok, db, err := NewDbConnection()
		if !ok {
			log.Fatal().Err(err).Msg("unable to connect to database")
		}

		stmt, err := db.Prepare(query)
		if err != nil {
			log.Error().Err(err).Msg("failed to prepare monitor watch query")
		}
		rows, err := stmt.Query()
		if err != nil {
			log.Error().Err(err).Msg("")
		}

		var description string
		var uniqueId string
		for rows.Next() {
			scanErr := rows.Scan(&description, &uniqueId)
			if scanErr != nil {
				log.Error().Err(scanErr).Msg("failed to get row")
			}

			obj := Monitor{
				UniqueId:    uniqueId,
				Description: description,
			}
			monitors <- []Monitor{obj}
		}
	}()
}
