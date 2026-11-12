package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2/google"
)

// fcmScope is the OAuth scope required to send Firebase Cloud Messaging v1
// messages.
const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

// fcmRequestTimeout bounds a single send. Token acquisition uses its own,
// slightly larger budget.
const fcmRequestTimeout = 10 * time.Second

type fcmMessage struct {
	Token        string            `json:"token"`
	Notification fcmNotification   `json:"notification"`
	Data         map[string]string `json:"data,omitempty"`
	Android      fcmAndroid        `json:"android"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type fcmAndroid struct {
	Notification fcmAndroidNotification `json:"notification"`
}

type fcmAndroidNotification struct {
	ChannelID string `json:"channel_id"`
}

type fcmSendBody struct {
	Message fcmMessage `json:"message"`
}

type fcmErrorBody struct {
	Error struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"error"`
}

// sendFCM delivers ev to each token over the FCM HTTP v1 API. Tokens Firebase
// reports as permanently invalid are passed to onInvalid so the caller can
// prune them; delivery failures are accumulated and returned without leaking
// token values.
func sendFCM(projectID, serviceAccountJSON string, tokens []string, ev Event, onInvalid func(token string)) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := google.JWTConfigFromJSON([]byte(serviceAccountJSON), fcmScope)
	if err != nil {
		return fmt.Errorf("invalid service account credentials: %w", err)
	}
	token, err := cfg.TokenSource(ctx).Token()
	if err != nil {
		return fmt.Errorf("could not obtain access token: %w", err)
	}

	data := make(map[string]string, len(ev.Data)+2)
	for key, value := range ev.Data {
		data[key] = value
	}
	data["type"] = ev.Type
	data["url"] = ev.URL

	url := "https://fcm.googleapis.com/v1/projects/" + projectID + "/messages:send"
	client := &http.Client{Timeout: fcmRequestTimeout}

	var failed int
	for _, deviceToken := range tokens {
		payload, err := json.Marshal(fcmSendBody{Message: fcmMessage{
			Token:        deviceToken,
			Notification: fcmNotification{Title: ev.Title, Body: ev.Body},
			Data:         data,
			Android: fcmAndroid{Notification: fcmAndroidNotification{
				ChannelID: "sandwitches_default",
			}},
		}})
		if err != nil {
			failed++
			continue
		}

		reqCtx, reqCancel := context.WithTimeout(context.Background(), fcmRequestTimeout)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			reqCancel()
			failed++
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		resp, err := client.Do(req)
		if err != nil {
			reqCancel()
			failed++
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		reqCancel()

		if resp.StatusCode == http.StatusOK {
			continue
		}
		failed++

		var apiErr fcmErrorBody
		_ = json.Unmarshal(body, &apiErr)
		status := apiErr.Error.Status
		if resp.StatusCode == http.StatusNotFound || status == "UNREGISTERED" ||
			(resp.StatusCode == http.StatusBadRequest && status == "INVALID_ARGUMENT") {
			if onInvalid != nil {
				onInvalid(deviceToken)
			}
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d message(s) rejected", failed, len(tokens))
	}
	return nil
}
