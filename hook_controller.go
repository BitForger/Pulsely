package main

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// HookCondition is a struct that represents the condition for a hook
// FailureThreshold is the number of failures before the hook is considered down
// DurationThreshold is the number of seconds before the hook is considered down
type HookCondition struct {
	FailureThreshold  int `json:"failureThreshold"`
	DurationThreshold int `json:"durationThreshold"`
}

// CreatHookBody is a struct that represents the request body for creating a hook
// Description is the description of the hook
// Condition is the condition for the hook
type CreatHookBody struct {
	Description string        `json:"description"`
	Condition   HookCondition `json:"condition"`
}

type UpdateHookBody struct {
	Description string        `json:"description"`
	Condition   HookCondition `json:"condition"`
}

func HookCreateHook(ctx fiber.Ctx) error {
	var bodyJson CreatHookBody
	err := json.Unmarshal(ctx.Body(), &bodyJson)
	if err != nil {
		log.Warn().Err(err).Msg("issue decoding body")
	}
	log.Info().Any("body", bodyJson.Description).Msg("print body")

	hs, _ := NewHookService()
	hook, _ := hs.CreateHook(bodyJson)

	return ctx.JSON(*hook)
}

func HookCreateHeartbeat(ctx fiber.Ctx) error {
	log.Info().Msg("Start request")
	id := ctx.Params("id")
	hSvc, _ := NewHookService()

	ok, err := hSvc.SaveHeartbeat(id, true)
	if ok {
		return ctx.Status(fiber.StatusOK).Next()
	}

	log.Error().Err(err).Msg("request failed")
	return ctx.SendStatus(fiber.StatusBadRequest)
}

func UpdateHeartbeat(ctx fiber.Ctx) error {
	var updateHookBody UpdateHookBody
	id := ctx.Params("id")
	err := json.Unmarshal(ctx.Body(), &updateHookBody)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode body")
	}

	hSvc, _ := NewHookService()
	ok, err := hSvc.UpdateHook(id, updateHookBody)
	if !ok {
		log.Error().Err(err).Msg("failed to update hook")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusOK)
}
