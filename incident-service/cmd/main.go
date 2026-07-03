package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/alert"
	"github.com/cQu1x/Incident-War-Room/internal/api"
	"github.com/cQu1x/Incident-War-Room/internal/auth"
	"github.com/cQu1x/Incident-War-Room/internal/bot"
	"github.com/cQu1x/Incident-War-Room/internal/config"
	"github.com/cQu1x/Incident-War-Room/internal/dashboard"
	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
	"github.com/cQu1x/Incident-War-Room/internal/mediastore"
	"github.com/cQu1x/Incident-War-Room/internal/reportclient"
	"github.com/cQu1x/Incident-War-Room/internal/repository"
	"github.com/cQu1x/Incident-War-Room/internal/service"
	"github.com/cQu1x/Incident-War-Room/internal/telegraphclient"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := repository.NewPool(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer pool.Close()

	incidents := repository.NewIncidentRepository(pool)
	events := repository.NewEventRepository(pool)
	txManager := repository.NewTxManager(pool)
	reports := reportclient.New(cfg.ReportServiceURL, reportclient.WithS3Enabled(cfg.S3Enabled))
	timelines := telegraphclient.New(telegraphclient.WithAccessToken(cfg.TelegraphAccessToken))

	var images media.Storage
	if cfg.S3Enabled {
		images = mediastore.New(mediastore.Config{
			EndpointURL:   cfg.S3EndpointURL,
			Region:        cfg.S3Region,
			Bucket:        cfg.S3Bucket,
			AccessKey:     cfg.S3AccessKey,
			SecretKey:     cfg.S3SecretKey,
			PublicBaseURL: cfg.S3PublicBaseURL,
		})
	}

	svc := service.New(incidents, events, txManager, reports, timelines, images)

	tgBot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalf("%v", errs.Wrapf(errs.KindUnavailable, "main", err, "connect to Telegram Bot API"))
	}

	tokens := auth.NewIssuer(cfg.JWTSecret, cfg.DashboardTokenTTL)
	dashboardLinker := dashboard.NewLinker(cfg.DashboardURL, tokens)

	handler := bot.New(
		svc,
		tgBot,
		bot.WithMediaEnabled(cfg.S3Enabled),
		bot.WithAlertChat(cfg.AlertChatID),
		bot.WithDashboard(dashboardLinker),
	)
	handler.Register(tgBot)

	alertWebhook := alert.NewHandler(handler, cfg.AlertmanagerWebhookToken)
	apiServer := api.NewServer(svc, tokens, cfg.CORSAllowedOrigin, api.Route{
		Pattern: "POST /webhooks/alertmanager",
		Handler: alertWebhook,
	})
	go func() {
		fmt.Printf("HTTP API listening on %s\n", cfg.HTTPAddr)
		if err := apiServer.Run(cfg.HTTPAddr); err != nil {
			log.Fatalf("http api: %v", err)
		}
	}()

	fmt.Println("Bot started")
	tgBot.Start()
}
