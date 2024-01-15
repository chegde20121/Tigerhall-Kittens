package service

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"math"
	"sync"

	"github.com/chegde20121/Tigerhall-Kittens/pkg/config"
	"github.com/chegde20121/Tigerhall-Kittens/pkg/messaging"
	log "github.com/sirupsen/logrus"
)

const (
	earthRadiusKm = 6371 // Radius of the Earth in kilometers
)

var (
	MessagingQueue *messaging.PubSub
	once           sync.Once
)

type Notifier struct {
	db     *sql.DB
	logger *log.Logger
}

func NewNotifier(db *sql.DB, logger *log.Logger) *Notifier {
	return &Notifier{db: db, logger: logger}
}

// haversineDistance calculates the Haversine distance between two points in kilometers.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Calculate differences
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	// Haversine formula
	a := math.Pow(math.Sin(deltaLat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(deltaLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers
	distance := earthRadiusKm * c

	return distance
}

type SightingNotification struct {
	TigerName     string
	TigerLocation string
	SightingTime  string
	userEmails    []string
}

func (n *Notifier) RegisterSightingSubscriber() chan any {
	n.logger.Info("Notifier initialized......")

	sightingServiceSubscriber := MessagingQueue.Subscribe()
	email := messaging.NewEmailHandler(n.logger)
	go func(subscriber chan any, name string) {
		for message := range subscriber {
			sightingNotification := message.(SightingNotification)
			data := messaging.EmailTemplateData{
				TigerName:     sightingNotification.TigerName,
				TigerLocation: sightingNotification.TigerLocation,
				SightingTime:  sightingNotification.SightingTime,
				Organization:  "Tigerhall-Kittens",
				ContactInfo:   fmt.Sprintf("Contact us at %v", config.GetEnvVar("SENDER_EMAIL")),
			}

			// Execute the email template
			emailBody, err := executeTemplate(data)
			if err != nil {
				n.logger.Error("Error executing email template:", err)
			} else {
				err := email.SendEmailNotification(sightingNotification.userEmails, "Tiger Sighting Notification", emailBody)
				if err != nil {
					n.logger.Error(err)
				}
			}

		}
		MessagingQueue.Unsubscribe(subscriber)
	}(sightingServiceSubscriber, "sightingServiceSubscriber")
	return sightingServiceSubscriber
}

// degreesToRadians converts degrees to radians.
func degreesToRadians(degrees float64) float64 {
	return degrees * (math.Pi / 180)
}

func GetMessagingQueue() *messaging.PubSub {
	once.Do(func() {
		MessagingQueue = messaging.NewPubSub()
	})
	return MessagingQueue
}

func executeTemplate(data messaging.EmailTemplateData) (string, error) {
	// Define the email template
	emailTemplate := `
Dear Sir/Madam,

We hope this message finds you well. We have exciting news to share with you regarding the tiger sighting you reported!

Recently, another sighting of the same tiger ({{.TigerName}}) has been reported by another user. Your dedication to wildlife observation and reporting is invaluable to our community.

Details of the recent sighting:
- Location: {{.TigerLocation}}
- Lastseen At: {{.SightingTime}}

We appreciate your commitment to tracking tiger populations in the wild. Your contributions play a crucial role in our conservation efforts.

Thank you for being an active member of our tiger tracking community.

Best regards,

{{.Organization}}
{{.ContactInfo}}
`

	// Create a new template and parse the email template
	tmpl, err := template.New("emailTemplate").Parse(emailTemplate)
	if err != nil {
		return "", err
	}

	// Execute the template with the provided data
	var emailBodyBuffer bytes.Buffer
	err = tmpl.Execute(&emailBodyBuffer, data)
	if err != nil {
		return "", err
	}

	return emailBodyBuffer.String(), nil
}
