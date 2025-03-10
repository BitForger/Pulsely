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
		failureThreshold INT DEFAULT 5,
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
	const createIntegrationsTable string = `
	CREATE TABLE IF NOT EXISTS integrations (
		id INTEGER NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		integrationType TEXT NOT NULL CHECK ( integrationType IN ('slack') ),
		botToken TEXT NOT NULL CHECK (integrationType = 'slack'),
		botUserId TEXT NOT NULL CHECK (integrationType = 'slack'),
		teamId TEXT NOT NULL CHECK (integrationType = 'slack'),
		enterpriseId TEXT NOT NULL CHECK (integrationType = 'slack')
	)`
	const createUsersTable string = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY,
		first_name TEXT	  NOT NULL,
		last_name TEXT	  NOT NULL,
		email TEXT	  NOT NULL,
		uid TEXT	  NOT NULL,
		role TEXT	  NOT NULL CHECK (role IN ('admin', 'user')),
		created_at DATETIME NOT NULL
	)`
	_, err := db.Exec(createMonitorsTable)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create table - monitors")
	}
	_, err = db.Exec(createHeartbeatsTable)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create table - heartbeats")
	}
	_, err = db.Exec(createIntegrationsTable)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create table - integrations")
	}
	_, err = db.Exec(createUsersTable)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create table - users")
	}
}
