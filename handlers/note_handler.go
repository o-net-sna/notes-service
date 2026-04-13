package handlers

import (
	"encoding/json"
	"net/http"
	"note-service/models"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type NoteHandler struct {
	repo *repository.NoteRepository
}

func NewNoteHandler(repo *repository.NoteRepository) *NoteHandler {
	return &NoteHandler{
		repo: repo,
	}
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "bad request")
		return
	}

	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "title is required")
		return
	}

	note, err := h.repo.Create(models.Note{
		Title:   input.Title,
		Content: input.Content,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"note": note,
	})
}

func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	notes, err := h.repo.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notes": notes,
	})
}

func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "note id is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid note id")
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		if err.Error() == "note not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NoteHandler) GetNoteByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "note id is required")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid note id")
		return
	}
	note, err := h.repo.GetByID(id)
	if err != nil {
		if err.Error() == "note not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"note": note,
	})

}
func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
