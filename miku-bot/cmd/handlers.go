package main

import (
	"context"
	"log/slog"
	"miku-bot/internal"
	"regexp"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

var socialLinkRe = regexp.MustCompile(`^https?://([a-z0-9-]+\.)*?(youtube\.com|youtu\.be|instagram\.com|tiktok\.com|twitter\.com|x\.com)(/\S*)?`)

func registerHandlers(bot *tele.Bot, workerClient *internal.Client, log *slog.Logger) {
	bot.Handle("/start", handleStart)
	bot.Handle(tele.OnText, makeHandleDownload(workerClient, log))
}

func handleStart(c tele.Context) error {
	return c.Send("👋 Привет! Я Мику. Скачаю для тебя видео из X, YouTube, TikTok, Inst и пришлю в чат. Могу скачать видео только весом не больше 50 мб")
}

func makeHandleDownload(workerClient *internal.Client, log *slog.Logger) tele.HandlerFunc {
	return func(c tele.Context) error {
		url, isExtracted := extractDownloadURL(c)

		if isExtracted == false {
			return nil
		}

		chatID := c.Chat().ID
		log.Info("download requested", "chat_id", chatID, "url", url)

		mess, err := c.Bot().Reply(c.Message(), "⏳ Начинаю загрузку")

		if err != nil {
			log.Error("reply failed", "chat_id", chatID, "error", err)
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		res := workerClient.Download(ctx, url)

		if res.Result == false {
			log.Warn("download failed", "chat_id", chatID, "url", url, "description", res.Description)
			_, editErr := c.Bot().Edit(mess, "❌ не получилось: "+res.Description)
			return editErr
		}

		log.Info("download succeeded", "chat_id", chatID, "url", url, "path", res.Path)

		_, err = c.Bot().Edit(mess, &tele.Video{File: tele.FromDisk(res.Path)})

		if err != nil {
			log.Error("send video failed", "chat_id", chatID, "path", res.Path, "error", err)
			_, editErr := c.Bot().Edit(mess, "❌ не получилось: "+err.Error())
			return editErr
		}

		return err
	}
}

func extractDownloadURL(c tele.Context) (string, bool) {
	mention := "@" + c.Bot().Me.Username
	if strings.Contains(c.Text(), mention) {

		if url := socialLinkRe.FindString(c.Text()); url != "" {
			return url, true
		}

		if c.Message().IsReply() {
			if url := socialLinkRe.FindString(c.Message().ReplyTo.Text); url != "" {
				return url, true
			}
		}
	} else {
		if url := socialLinkRe.FindString(c.Text()); url != "" && c.Chat().Type == tele.ChatPrivate {
			return url, true
		}
	}

	return "", false
}
