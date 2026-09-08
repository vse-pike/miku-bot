package main

import (
	"context"
	"log/slog"
	"miku-bot/internal"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if err := godotenv.Load(); err != nil {
		log.Error("no .env file found, relying on real env vars")
		os.Exit(1)
	}

	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Error("BOT_TOKEN is required")
		os.Exit(1)
	}

	pyWorkerUrl := os.Getenv("WORKER_URL")

	if pyWorkerUrl == "" {
		log.Error("WORKER_URL is required")
		os.Exit(1)
	}

	downloaderClient := internal.CreateClient(pyWorkerUrl, log)

	bot, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Error("create bot", "error", err)
		os.Exit(1)
	}

	bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			start := time.Now()
			err := next(c)
			log.Info("update handled",
				"from", c.Sender().Username,
				"text", c.Text(),
				"took", time.Since(start))
			return err
		}
	})

	registerHandlers(bot, downloaderClient, log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Info("shutting down")
		bot.Stop()
	}()

	log.Info("bot started")
	bot.Start()
}
