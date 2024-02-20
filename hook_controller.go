package main

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type CreatHookBody struct {
	Description string `json:"description"`
}

func HookCreatehook(ctx fiber.Ctx) error {
	var bodyJson CreatHookBody
	err := json.Unmarshal(ctx.Body(), &bodyJson)
	if err != nil {
		log.Warn().Err(err).Msg("issue decoding body")
	}
	log.Info().Any("body", bodyJson.Description).Msg("print body")

	hs, _ := NewHookService()
	hook, _ := hs.CreateHook(bodyJson.Description)

	return ctx.JSON(*hook)
}

func HookCreateHeartbeat(ctx fiber.Ctx) error {
	return nil
}
