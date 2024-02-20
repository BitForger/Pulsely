package main

import (
	"crypto/sha512"
	"database/sql"
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
	hostEnvVar, _ := os.LookupEnv("BEATMON_HOST")
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

	const createDb string = `
      	CREATE TABLE IF NOT EXISTS monitors (
      	id INTEGER NOT NULL PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		description TEXT,
		uniqueId TEXT
		)`
	log.Debug().Msg("Create DB if doesn't exist")
	_, err = db.Exec(createDb)
	if err != nil {
		log.Error().Err(err).Msg("unable to create table")
		return HookService{}, err
	}

	return HookService{
		Host:      hostEnvVar,
		TokenSalt: tokenSalt,
		Db:        db,
	}, nil
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
	_, err = stmt.Exec(time.Now().Unix(), description, uniqueId.String())
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
