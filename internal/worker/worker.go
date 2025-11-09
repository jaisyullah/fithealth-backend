package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jaisyullah/fithealth-backend/internal/config"
	"github.com/jaisyullah/fithealth-backend/internal/mapper"
	"github.com/jaisyullah/fithealth-backend/internal/oauth"
	"github.com/jaisyullah/fithealth-backend/internal/store"

	"github.com/redis/go-redis/v9"
)

type SenderWorker struct {
	DB      *store.DB
	Redis   *redis.Client
	Token   *oauth.TokenManager
	Cfg     *config.Config
	client  *http.Client
}

func NewSenderWorker(db *store.DB, r *redis.Client, t *oauth.TokenManager, cfg *config.Config) *SenderWorker {
	return &SenderWorker{
		DB:     db,
		Redis:  r,
		Token:  t,
		Cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (w *SenderWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// pop job
			job, err := w.Redis.LPop(ctx, "obs_queue").Result()
			if err == redis.Nil {
				time.Sleep(time.Duration(w.Cfg.MaxWorkerPollInterval) * time.Second)
				continue
			}
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			// parse id
			id, err := strconv.ParseInt(job, 10, 64)
			if err != nil {
				continue
			}
			obs, err := w.DB.GetObservationByID(ctx, id)
			if err != nil {
				continue
			}
			// map to FHIR
			fhirBytes, err := mapper.ToFHIRObservation(obs.PatientID, obs.ObsType, obs.Value, obs.Unit, obs.ObservedAt)
			if err != nil {
				_ = w.DB.IncrementRetry(ctx, obs.ID, err.Error())
				continue
			}
			// get token
			token, err := w.Token.GetToken()
			if err != nil {
				_ = w.DB.IncrementRetry(ctx, obs.ID, err.Error())
				continue
			}
			req, _ := http.NewRequest("POST", w.Cfg.SatusehatFHIRUrl+"/Observation", bytesReader(fhirBytes))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			resp, err := w.client.Do(req)
			if err != nil {
				_ = w.DB.IncrementRetry(ctx, obs.ID, err.Error())
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			// save transaction
			ft := &store.FHIRTransaction{
				ObservationRawID: obs.ID,
				FHIRPayload:      string(fhirBytes),
				ResponseCode:     resp.StatusCode,
				ResponseBody:     string(body),
			}
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				ft.Status = "success"
				_ = w.DB.SaveFHIRTransaction(ctx, ft)
				_ = w.DB.MarkSent(ctx, obs.ID, resp.StatusCode)
			} else {
				ft.Status = "failed"
				_ = w.DB.SaveFHIRTransaction(ctx, ft)
				_ = w.DB.IncrementRetry(ctx, obs.ID, fmt.Sprintf("status:%d body:%s", resp.StatusCode, string(body)))
			}
		}
	}
}

func bytesReader(b []byte) *bytesReaderType {
	return &bytesReaderType{b: b}
}

type bytesReaderType struct {
	b []byte
	i int64
}

func (r *bytesReaderType) Read(p []byte) (n int, err error) {
	if r.i >= int64(len(r.b)) {
		return 0, io.EOF
	}
	n = copy(p, r.b[r.i:])
	r.i += int64(n)
	return
}

func (r *bytesReaderType) Close() error { return nil }
