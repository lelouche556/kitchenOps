package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"SwishAssignment/internal/application"
	"github.com/go-chi/chi/v5"
)

type KitchenController struct {
	service *application.KitchenAppService
}

var (
	kitchenControllerOnce sync.Once
	kitchenControllerInst *KitchenController
)

func NewKitchenController(service *application.KitchenAppService) *KitchenController {
	kitchenControllerOnce.Do(func() {
		kitchenControllerInst = &KitchenController{service: service}
	})
	return kitchenControllerInst
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseTaskID(r *http.Request) (uint64, error) {
	return strconv.ParseUint(chi.URLParam(r, "taskID"), 10, 64)
}

func (c *KitchenController) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
