package main

import (
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/pbkdf2"
	_ "modernc.org/sqlite"
	"modernc.org/strutil"
	"os"
	"time"
)

type HookService struct {
	Host      string
	TokenSalt string
	Db        *sql.DB
}

type CreatedHook struct {
	Hook        string `json:"hook"`
	Description string `json:"description"`
	Token       string `json:"token"`
}

func NewHookService() (HookService, error) {
	hostEnvVar, found := os.LookupEnv("BEATMON_HOST")
	if !found {
		hostEnvVar = "localhost"
	}
	tokenSalt, _ := os.LookupEnv("BEATMON_TOKEN_SALT")
	dbString, _ := os.LookupEnv("BEATMON_SQLITE_FILE_LOCATION")
	if dbString == "" {
		dbString, _ = os.UserHomeDir()
		dbString += "/monitors.db"
	}
	log.Info().Any("DB connection string", dbString).Msg("Opening DB connection")
	db, err := sql.Open("sqlite", dbString)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open DB connection")
		return HookService{}, err
	}

	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to ping db connection")
		return HookService{}, err
	}

	prepareTables(db)

	return HookService{
		Host:      hostEnvVar,
		TokenSalt: tokenSalt,
		Db:        db,
	}, nil
}

func prepareTables(db *sql.DB) {
	log.Debug().Msg("Create tables if doesn't exist")
	const createMonitorsTable string = `
	CREATE TABLE IF NOT EXISTS monitors (
		id INTEGER NOT NULL PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		description TEXT,
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

func (s *HookService) CreateHook(description string) (*CreatedHook, error) {
	if s == nil {
		log.Error().Msg("s is nil")
	}
	if &s.Db == nil {
		log.Warn().Msg("DB is nil")
	}
	const insertQuery = `INSERT INTO monitors (timestamp, description, uniqueId) VALUES (?, ?, ?)`
	stmt, err := s.Db.Prepare(insertQuery)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, err
	}
	defer stmt.Close()

	uniqueId := ksuid.New()
	_, err = stmt.Exec(time.Now().UTC(), description, uniqueId.String())
	if err != nil {
		return nil, err
	}

	return &CreatedHook{
		Hook:        fmt.Sprintf("https://%s/hooks/%s", s.Host, uniqueId.String()),
		Description: description,
		Token:       s.GetToken(uniqueId.String()),
	}, nil
}

func (s *HookService) GetToken(id string) string {
	dk := pbkdf2.Key([]byte(id), []byte(s.TokenSalt), 1024, 32, sha512.New)
	return string(strutil.Base64Encode(dk))
}

func (s *HookService) SaveHeartbeat(id string, up bool) (ok bool, err error) {
	// Check that hook exists
	const findQuery string = "SELECT EXISTS(SELECT 1 FROM monitors WHERE uniqueId=?)"
	stmt, err := s.Db.Prepare(findQuery)
	if err != nil {
		log.Error().Err(err).Msg("failed to prepare find statement")
		return false, err
	}
	defer stmt.Close()

	res, stmtErr := stmt.Exec(id)
	if stmtErr != nil {
		errMsg := fmt.Sprintf("error getting rows for id - %s", id)
		log.Error().Err(stmtErr).Msg(errMsg)
		return false, errors.New(errMsg)
	}
	affectedRows, _ := res.RowsAffected()
	if affectedRows < 1 {
		errMsg := fmt.Sprintf("no monitor found for id - %s", id)
		log.Error().Msg(errMsg)
		return false, errors.New(errMsg)
	}

	// insert heartbeat status
	const insertQuery string = "INSERT INTO heartbeats (timestamp, up, hookId) VALUES (?,?,?)"
	stmt, err = s.Db.Prepare(insertQuery)
	if err != nil {
		log.Error().Err(err).Msg("failed to prepare insert statement")
		return false, err
	}
	defer stmt.Close()

	var upStatus = 0
	if up {
		upStatus = 1
	}

	_, err = stmt.Exec(time.Now().Unix(), upStatus, id)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert heartbeat status")
		return false, err
	}

	return true, nil
}

func (s *HookService) UpdateHook(id, description string) (bool, error) {
	if s == nil {
		log.Error().Msg("s is nil")
	}
	if &s.Db == nil {
		log.Warn().Msg("DB is nil")
	}
	defer s.Db.Close()
	const query = `UPDATE monitors SET description = ? WHERE uniqueId = ?`
	stmt, err := s.Db.Prepare(query)
	if err != nil {
		log.Error().Err(err).Msg("failed to prepare update query")
		return false, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(description, id)
	if err != nil {
		log.Error().Err(err).Err(err).Msg("failed to update hook")
		return false, err
	}
	rowsAff, _ := result.RowsAffected()
	if rowsAff < 1 {
		log.Warn().Any("rows affected", rowsAff).Msg("no error received but no rows affected")
	}

	return true, nil
}
