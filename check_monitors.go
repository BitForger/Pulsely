package main

import (
	"github.com/rs/zerolog/log"
)

// Get all monitors
// Spawn goroutine to watch for new monitors
// Spawn goroutines for each monitor to check the heartbeat status
func start() {
	const query = `SELECT * FROM monitors`
	ok, db, err := NewDbConnection()
	if !ok {
		log.Fatal().Err(err).Msg("unable to connect to database")
	}

	go func() {

	}()
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Error().Err(err).Msg("failed to prepare monitor watch query")
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	var row any
	for rows.Next() {
		scanErr := rows.Scan(&row)
		if scanErr != nil {
			log.Error().Err(scanErr).Msg("failed to get row")
		}
	}

}
