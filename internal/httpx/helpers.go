package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/utils"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func PathID(r *http.Request, key string) (int, error) {
	id := utils.ParseInt(r.PathValue(key), -1)
	if id < 0 {
		return 0, fmt.Errorf("invalid id %q: %w", r.PathValue("id"), apperrors.ErrInvalidInput)
	}
	return id, nil
}
