package notifier

import (
	"context"
	"log"

	"github.com/aouxes/uptime-monitor/internal/storage"
	"github.com/aouxes/uptime-monitor/internal/telegram"
)

type Notifier struct {
	telegram *telegram.Client
	storage  *storage.Storage
}

func New(telegramToken string, storage *storage.Storage) *Notifier {
	return &Notifier{
		telegram: telegram.NewClient(telegramToken),
		storage:  storage,
	}
}

func (n *Notifier) NotifySiteStatusChange(ctx context.Context, siteID int, oldStatus, newStatus string) error {
	// Получаем информацию о сайте и пользователе
	site, err := n.storage.GetSiteByID(ctx, siteID)
	if err != nil {
		return err
	}

	if site == nil {
		log.Printf("Site with ID %d not found", siteID)
		return nil
	}

	// Получаем информацию о пользователе
	user, err := n.storage.GetUserByID(ctx, site.UserID)
	if err != nil {
		return err
	}

	if user == nil || user.TelegramChatID == 0 {
		log.Printf("User %d has no Telegram chat ID configured", site.UserID)
		return nil
	}

	// Отправляем уведомление
	return n.telegram.SendSiteStatusNotification(ctx, user.TelegramChatID, site.URL, oldStatus, newStatus)
}
