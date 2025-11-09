package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jaisyullah/fithealth-backend/internal/config"
	"github.com/jaisyullah/fithealth-backend/internal/oauth"
	"github.com/jaisyullah/fithealth-backend/internal/store"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type ServerCtx struct {
	DB       *store.DB
	Redis    *redis.Client
	TokenMgr *oauth.TokenManager
	Cfg      *config.Config
}

func NewServer(db *store.DB, r *redis.Client, t *oauth.TokenManager, cfg *config.Config) *echo.Echo {
	e := echo.New()

	ctx := &ServerCtx{DB: db, Redis: r, TokenMgr: t, Cfg: cfg}
	e.POST("/v1/observations", ctx.PostObservations)
	e.GET("/health", ctx.Health)

	return e
}

type incomingObs struct {
	DeviceID    string `json:"deviceId"`
	PatientID   string `json:"patientId"`
	ObsType     string `json:"type"`
	Value       string `json:"value"` // accept string or number
	Unit        string `json:"unit"`
	Timestamp   string `json:"timestamp"` // ISO8601 local
}

// Request body supports batch
type obsRequest struct {
	DeviceID     string        `json:"deviceId"`
	PatientID    string        `json:"patientId"`
	Observations []incomingObs `json:"observations"`
}

func (s *ServerCtx) PostObservations(c echo.Context) error {
	var req obsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	ctx := c.Request().Context()
	for _, o := range req.Observations {
		// parse value
		vf := 0.0
		if o.Value != "" {
			if valFloat, err := strconv.ParseFloat(o.Value, 64); err == nil {
				vf = valFloat
			}
		}
		// parse timestamp
		t := time.Now()
		if o.Timestamp != "" {
			if parsed, err := time.Parse(time.RFC3339, o.Timestamp); err == nil {
				t = parsed
			}
		}

		or := &store.ObservationRaw{
			DeviceID:   req.DeviceID,
			PatientID:  req.PatientID,
			ObsType:    o.ObsType,
			Value:      vf,
			Unit:       o.Unit,
			ObservedAt: t.UTC(),
			Status:     "pending",
		}
		if err := s.DB.CreateObservation(ctx, or); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}
		// push job id to redis queue
		_ = s.Redis.RPush(ctx, "obs_queue", or.ID).Err()
		// mark queued
		_ = s.DB.UpdateStatus(ctx, or.ID, "queued", "")
	}
	return c.JSON(http.StatusAccepted, map[string]string{"status": "queued"})
}

func (s *ServerCtx) Health(c echo.Context) error {
	ok := "ok"
	_ = s.DB.VerifyConnection()
	return c.JSON(http.StatusOK, map[string]string{"status": ok})
}
