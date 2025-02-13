package apperrors

import (
	"avito-backend-intern-winter25/pkg/errs"
	"errors"
	"net/http"
)

func ToResponse(err error) (int, interface{}) {
	switch {
	case errors.Is(err, errs.ErrInvalidInput):
		return http.StatusBadRequest, ErrorResponse{Error: "Неверный запрос"}
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, ErrorResponse{Error: "Неавторизован"}
	default:
		return http.StatusInternalServerError, ErrorResponse{Error: "Внутренняя ошибка сервера"}
	}
}
