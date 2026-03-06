package apierror

import (
	"errors"
	"fmt"
	"net/http"
)

type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Status  int               `json:"-"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

const (
	FieldFirstName  = "first_name"
	FieldMiddleName = "middle_name"
	FieldLastName   = "last_name"
	FieldEmail      = "email"
	FieldPassword   = "password"
	FieldBirthDate  = "birth_date"
	FieldCode       = "code"
)

const (
	ErrInvalidFirstName  = "Некорректно указано имя"
	ErrInvalidMiddleName = "Некорректно указано отчество"
	ErrInvalidLastName   = "Некорректно указана фамилия"
	ErrInvalidEmail      = "Некорректный адрес электронной почты"
	ErrInvalidBirthDate  = "Некорректная дата рождения. Регистрация доступна для лиц, достигших 14 лет"
	ErrInvalidPassword   = "Пароль должен состоять из 8 и более символов и содержать как минимум одну латинскую букву и одну цифру"
)

var (
	ErrInternal = &APIError{
		Code:    "internal_error",
		Message: "Что-то пошло не так. Повторите попытку позже",
		Status:  http.StatusInternalServerError,
	}

	ErrInvalidJSON = &APIError{
		Code:    "invalid_json",
		Message: "Некорректный формат JSON",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidVerificationCode = &APIError{
		Code:    "invalid_verification_code",
		Message: "Неверный код подтверждения",
		Status:  http.StatusBadRequest,
	}

	ErrCodeExpired = &APIError{
		Code:    "code_expired",
		Message: "Срок действия кода истёк",
		Status:  http.StatusBadRequest,
	}

	ErrEmailIsUsed = &APIError{
		Code:    "email_already_used",
		Message: "Пользователь с таким email уже существует",
		Status:  http.StatusConflict,
	}

	ErrInvalidCredentials = &APIError{
		Code:    "invalid_credentials",
		Message: "Неверный email или пароль",
		Status:  http.StatusUnauthorized,
	}

	ErrUnauthorized = &APIError{
		Code:    "unauthorized",
		Message: "Требуется авторизация",
		Status:  http.StatusUnauthorized,
	}

	ErrInvalidToken = &APIError{
		Code:    "invalid_token",
		Message: "Неверный или просроченный токен",
		Status:  http.StatusUnauthorized,
	}

	ErrInvalidTokenFormat = &APIError{
		Code:    "invalid_token_format",
		Message: "Ожидается Bearer <token>",
		Status:  http.StatusUnauthorized,
	}

	ErrTooManyAttempts = &APIError{
		Code:    "too_many_attempts",
		Message: "Превышено количество попыток ввода кода подтверждения. Попробуйте позже",
		Status:  http.StatusTooManyRequests,
	}

	ErrCodeRateLimited = &APIError{
		Code:    "rate_limit_exceeded",
		Message: "Превышено количество запросов кода подтверждения. Попробуйте позже",
		Status:  http.StatusTooManyRequests,
	}
)

func New(code, message string, status int) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

func Wrap(err error, fallback *APIError) *APIError {
	if err == nil {
		return nil
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	switch {
	case errors.Is(err, ErrInternal):
		return ErrInternal
	}

	return fallback
}

func Validation(fields map[string]string) *APIError {
	e := &APIError{
		Code:    "validation_failed",
		Message: "Ошибка при указании данных",
		Status:  http.StatusUnprocessableEntity,
		Fields:  fields,
	}
	if len(fields) == 1 {
		for _, msg := range fields {
			e.Message = msg
		}
	}
	return e
}
