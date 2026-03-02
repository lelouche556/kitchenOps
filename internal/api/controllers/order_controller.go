package controllers

import (
	"encoding/json"
	"net/http"

	"SwishAssignment/internal/application"
)

func (c *KitchenController) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	var req application.ConfirmOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	orderID, ready, err := c.service.ConfirmOrder(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"order_id": orderID, "ready_task_ids": ready})
}

func (c *KitchenController) AllocateOnce(w http.ResponseWriter, r *http.Request) {
	taskID, err := c.service.AllocateOnce(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assigned_task_id": taskID})
}
