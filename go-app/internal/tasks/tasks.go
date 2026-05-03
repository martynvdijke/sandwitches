package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"github.com/martynvdijke/sandwitches-go/internal/config"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

var cfg *config.Config

func Init(c *config.Config) {
	cfg = c
}

func EnqueueNotifyOrder(orderID uint) {
	go func() {
		time.Sleep(100 * time.Millisecond)
		notifyOrderSubmitted(orderID)
	}()
}

func EnqueueEmailUsers(recipeID uint) {
	go func() {
		time.Sleep(100 * time.Millisecond)
		notifyNewRecipe(recipeID)
	}()
}

func EnqueueGotify(title, message string, priority int) {
	go func() {
		sendGotify(title, message, priority)
	}()
}

func EnqueueResetDailyOrders() {
	go func() {
		database.DB.Model(&database.Recipe{}).Where("1 = 1").Update("daily_orders_count", 0)
		log.Println("Daily order counts reset")
	}()
}

func notifyOrderSubmitted(orderID uint) {
	var order database.Order
	if err := database.DB.Preload("Items.Recipe").Preload("User").First(&order, orderID).Error; err != nil {
		log.Printf("Order #%d not found for notification", orderID)
		return
	}

	if order.User.Email == "" {
		return
	}

	var itemsList string
	for _, item := range order.Items {
		itemsList += fmt.Sprintf("- %dx %s\n", item.Quantity, item.Recipe.Title)
	}

	subject := fmt.Sprintf("Order Confirmation: Order #%d", order.ID)
	body := fmt.Sprintf(`Hello %s,

Your order has been submitted!

Order ID: %d
Items:
%s
Total: %.2f EUR

Track your order: %s/orders/track/%s

Thank you for ordering with Sandwitches.`,
		order.User.Username, order.ID, itemsList, order.TotalPrice, baseURL(), order.TrackingToken)

	sendEmail(order.User.Email, subject, body)

	EnqueueGotify("New Order Received",
		fmt.Sprintf("Order #%d by %s. Total: %.2f EUR", order.ID, order.User.Username, order.TotalPrice), 6)
}

func notifyNewRecipe(recipeID uint) {
	var recipe database.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		return
	}

	var emails []string
	database.DB.Model(&database.User{}).Where("email != ''").Pluck("email", &emails)

	var uploader string
	if recipe.UploadedByID != nil {
		var user database.User
		if err := database.DB.First(&user, *recipe.UploadedByID).Error; err == nil {
			uploader = user.Username
		}
	}

	subject := fmt.Sprintf("New Recipe: %s by %s", recipe.Title, uploader)
	url := fmt.Sprintf("%s/recipes/%s", baseURL(), recipe.Slug)

	htmlBody := fmt.Sprintf(`<div style="font-family:sans-serif;max-width:600px;margin:auto;border:1px solid #eee;padding:20px">
<h2 style="color:#d35400;text-align:center">New Recipe: %s by %s</h2>
<hr><p>%s</p>
<p>Full recipe: <a href="%s">%s</a></p>
</div>`, recipe.Title, uploader, recipe.Description, url, url)

	textBody := fmt.Sprintf("New recipe: %s\n\n%s\n\nView: %s", recipe.Title, recipe.Description, url)

	for _, email := range emails {
		sendHTMLEmail(email, subject, textBody, htmlBody)
	}

	EnqueueGotify("New Recipe Uploaded",
		fmt.Sprintf("'%s' uploaded by %s.", recipe.Title, uploader), 5)
}

func sendGotify(title, msg string, priority int) {
	url := ""
	token := ""

	var setting database.Setting
	database.DB.First(&setting)
	url = setting.GotifyURL
	token = setting.GotifyToken

	if url == "" || token == "" {
		return
	}

	payload := map[string]interface{}{
		"title":    title,
		"message":  msg,
		"priority": priority,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(fmt.Sprintf("%s/message?token=%s", url, token),
		"application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("Gotify notification failed: %v", err)
		return
	}
	resp.Body.Close()
}

func sendEmail(to, subject, body string) {
	if cfg == nil || !cfg.SendEmail {
		return
	}
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		cfg.SMTPFromEmail, to, subject, body)
	err := smtp.SendMail(fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort),
		auth, cfg.SMTPFromEmail, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("Email to %s failed: %v", to, err)
	}
}

func sendHTMLEmail(to, subject, textBody, htmlBody string) {
	if cfg == nil || !cfg.SendEmail {
		return
	}
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	boundary := "boundary-42"
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n\r\n--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n--%s--",
		cfg.SMTPFromEmail, to, subject, boundary, boundary, textBody, boundary, htmlBody, boundary)
	err := smtp.SendMail(fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort),
		auth, cfg.SMTPFromEmail, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("HTML Email to %s failed: %v", to, err)
	}
}

func baseURL() string {
	return "http://localhost"
}
