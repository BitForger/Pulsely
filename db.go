package main

import (
	"database/sql"
	"github.com/rs/zerolog/log"
	"os"
)

func NewDbConnection() (bool, *sql.DB, error) {
	dbString, _ := os.LookupEnv("BEATMON_SQLITE_FILE_LOCATION")
	if dbString == "" {
		dbString, _ = os.UserHomeDir()
		dbString += "/monitors.db"
	}
	log.Info().Any("DB connection string", dbString).Msg("Opening DB connection")
	db, err := sql.Open("sqlite", dbString)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open DB connection")
		return false, nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to ping db connection")
		return false, nil, err
	}

	prepareTables(db)

	return true, db, nil
}

func prepareTables(db *sql.DB) {
	log.Debug().Msg("Create tables if doesn't exist")
	const createMonitorsTable string = `
	CREATE TABLE IF NOT EXISTS monitors (
		id INTEGER NOT NULL PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		description TEXT,
		failureThreshold TEXT DEFAULT 5,
		durationThreshold TEXT DEFAULT '5m',
		uniqueId TEXT
	)`
	const createHeartbeatsTable string = `
	CREATE TABLE IF NOT EXISTS heartbeats (
		id INTEGER NOT NULL PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		up INTEGER NOT NULL CHECK (up IN (0, 1)),
		hookId TEXT NOT NULL
	)`
	_, err := db.Exec(createMonitorsTable)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create table - monitors")
	}
	_, err = db.Exec(createHeartbeatsTable)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create table - heartbeats")
	}
}
