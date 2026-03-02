package controllers

import (
	"net/http"

	"SwishAssignment/internal/repository"
)

func (c *KitchenController) StartTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseTaskID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid taskID"})
		return
	}
	if err := c.service.StartTask(r.Context(), taskID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "STARTED"})
}

func (c *KitchenController) CompleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseTaskID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid taskID"})
		return
	}
	if err := c.service.CompleteTask(r.Context(), taskID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "COMPLETED"})
}

func (c *KitchenController) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseTaskID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid taskID"})
		return
	}
	task, err := c.service.GetTask(r.Context(), taskID)
	if err != nil {
		if err == repository.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, task)
}
