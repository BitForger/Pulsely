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

	return true, db, nil
}
