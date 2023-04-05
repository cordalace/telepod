package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"codeberg.org/cordalace/telepod/internal/podruntime"
	"codeberg.org/cordalace/telepod/internal/telegramnotifier"
	"codeberg.org/cordalace/telepod/internal/versionsdb"
	"codeberg.org/cordalace/telepod/internal/workflow"
)

func requireEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("error missing required environment variable: %s", name)
	}

	return value
}

func main() {
	telegramChatID := requireEnv("TELEGRAM_CHAT_ID")
	telegramBotToken := requireEnv("TELEGRAM_BOT_TOKEN")

	podRuntime := podruntime.NewPodRuntime()

	versionsDB := versionsdb.NewVersionsDB()
	if err := versionsDB.Init(); err != nil {
		log.Fatalf("error initializing versions db: %v", err)
	}

	httpClient := new(http.Client)

	telegramNotifier := telegramnotifier.NewTelegramNotifier(httpClient, telegramChatID, telegramBotToken)

	wf := workflow.NewWorkflow(podRuntime, versionsDB, telegramNotifier)

	if err := wf.Run(context.Background()); err != nil {
		log.Fatalf("error running workflow: %v", err)
	}
}
