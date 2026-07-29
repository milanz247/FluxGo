package handlers

import (
	"net/http"

	Route "fluxgo/internal/route"
	"gorm.io/gorm"
)

func Health(database *gorm.DB) Route.Handler {
	return func(c *Route.Context) error {
		sqlDB, err := database.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			return c.JSON(http.StatusServiceUnavailable, Route.Data{"status": "unavailable"})
		}
		return c.OK(Route.Data{"status": "ok"})
	}
}
