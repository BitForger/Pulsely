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

// CreatedHook is a struct that represents the response body from creating a hook
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

	ok, db, err := NewDbConnection()
	if !ok {
		log.Fatal().Err(err).Msg("unable to connect to database")
	}
	return HookService{
		Host:      hostEnvVar,
		TokenSalt: tokenSalt,
		Db:        db,
	}, nil
}

func (s *HookService) CreateHook(body CreatHookBody) (*CreatedHook, error) {
	if s == nil {
		log.Error().Msg("s is nil")
	}
	if &s.Db == nil {
		log.Warn().Msg("DB is nil")
	}
	const insertQuery = `
	INSERT INTO monitors (timestamp, description, uniqueId, failureThreshold, durationThreshold) 
		VALUES (?, ?, ?, ?, ?)
	`
	stmt, err := s.Db.Prepare(insertQuery)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, err
	}
	defer stmt.Close()

	uniqueId := ksuid.New()
	_, err = stmt.Exec(time.Now().UTC(),
		body.Description,
		uniqueId.String(),
		body.Condition.FailureThreshold,
		body.Condition.DurationThreshold,
	)
	if err != nil {
		return nil, err
	}

	return &CreatedHook{
		Hook:        fmt.Sprintf("https://%s/hooks/%s", s.Host, uniqueId.String()),
		Description: body.Description,
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

func (s *HookService) UpdateHook(id string, body UpdateHookBody) (bool, error) {
	if s == nil {
		log.Error().Msg("s is nil")
	}
	if &s.Db == nil {
		log.Warn().Msg("DB is nil")
	}
	defer s.Db.Close()

	// Update description
	if body.Description != "" {
		const updateDescriptionQuery = `UPDATE monitors SET description = ? WHERE uniqueId = ?`
		stmt, err := s.Db.Prepare(updateDescriptionQuery)
		if err != nil {
			log.Error().Err(err).Msg("failed to prepare update query")
			return false, err
		}
		defer stmt.Close()

		result, err := stmt.Exec(body.Description, id)
		if err != nil {
			log.Error().Err(err).Err(err).Msg("failed to update hook")
			return false, err
		}
		rowsAff, _ := result.RowsAffected()
		if rowsAff < 1 {
			log.Warn().Any("rows affected", rowsAff).Msg("no error received but no rows affected")
		}
	}

	// Update condition
	if body.Condition != (HookCondition{}) {
		const updateDurationQuery = `UPDATE monitors SET durationThreshold = ?, failureThreshold = ? WHERE uniqueId = ?`
		stmt, err := s.Db.Prepare(updateDurationQuery)
		if err != nil {
			log.Error().Err(err).Msg("failed to prepare update query")
			return false, err
		}
		defer stmt.Close()

		result, err := stmt.Exec(body.Condition.DurationThreshold, body.Condition.FailureThreshold, id)
		if err != nil {
			log.Error().Err(err).Msg("failed to update hook")
			return false, err
		}
		rowsAff, _ := result.RowsAffected()
		if rowsAff < 1 {
			log.Warn().Any("rows affected", rowsAff).Msg("no error received but no rows affected")
		}
	}

	return true, nil
}
