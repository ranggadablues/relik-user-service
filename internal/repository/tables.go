package repository

import (
	"relik-user-service/internal/models"
)

func UserCollection() string {
	return new(models.User).Collection()
}

func UserDetailsCollection() string {
	return new(models.UserDetails).Collection()
}

func NotificationIntegrationsCollection() string {
	return new(models.NotificationIntegrations).Collection()
}

func NotificationEventsCollection() string {
	return new(models.NotificationEvents).Collection()
}
