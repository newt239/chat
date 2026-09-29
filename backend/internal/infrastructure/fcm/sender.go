package fcm

import (
	"context"
	"errors"
	"fmt"
	"maps"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"

	"github.com/newt239/chat/internal/domain/entity"
	notificationuc "github.com/newt239/chat/internal/usecase/notification"
)

// FCM の 1 回の SendEach で送れる上限
const batchSize = 500

// Sender は Firebase Admin SDK で FCM へ通知を送ります。認証情報は ADC から取る
type Sender struct {
	client *messaging.Client
}

func NewSender(ctx context.Context, projectID string) (*Sender, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize messaging client: %w", err)
	}
	return &Sender{client: client}, nil
}

func (s *Sender) Send(ctx context.Context, messages []notificationuc.PushMessage) ([]string, error) {
	invalid := []string{}
	var errs []error
	for start := 0; start < len(messages); start += batchSize {
		batch := messages[start:min(start+batchSize, len(messages))]
		converted := make([]*messaging.Message, 0, len(batch))
		for _, m := range batch {
			converted = append(converted, toFCMMessage(m))
		}
		res, err := s.client.SendEach(ctx, converted)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for i, r := range res.Responses {
			switch {
			case r.Success:
			case messaging.IsUnregistered(r.Error) || messaging.IsSenderIDMismatch(r.Error):
				invalid = append(invalid, batch[i].Token)
			default:
				errs = append(errs, r.Error)
			}
		}
	}
	return invalid, errors.Join(errs...)
}

// toFCMMessage はウェブにはデータだけを送る。表示は Service Worker が決め、前面にいるときは出さないため
func toFCMMessage(m notificationuc.PushMessage) *messaging.Message {
	msg := &messaging.Message{Fid: m.Token, Data: m.Data}
	switch m.Platform {
	case entity.PushPlatformWeb:
		data := maps.Clone(m.Data)
		data["title"] = m.Title
		data["body"] = m.Body
		msg.Data = data
	case entity.PushPlatformAndroid:
		msg.Android = &messaging.AndroidConfig{
			Priority:     "high",
			Notification: &messaging.AndroidNotification{Title: m.Title, Body: m.Body, Tag: m.Data["channelId"]},
		}
	case entity.PushPlatformIOS:
		msg.APNS = &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{Aps: &messaging.Aps{
				Alert:    &messaging.ApsAlert{Title: m.Title, Body: m.Body},
				ThreadID: m.Data["channelId"],
				Sound:    "default",
			}},
		}
	}
	return msg
}
