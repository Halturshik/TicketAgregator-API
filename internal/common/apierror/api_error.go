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
	FieldFirstName    = "first_name"
	FieldMiddleName   = "middle_name"
	FieldLastName     = "last_name"
	FieldEmail        = "email"
	FieldPassword     = "password"
	FieldBirthDate    = "birth_date"
	FieldPassengers   = "passengers"
	FieldDocument     = "document"
	FieldCode         = "code"
	FieldRefreshToken = "refresh_token"
)

const (
	ErrInvalidFirstName    = "Некорректно указано имя"
	ErrInvalidMiddleName   = "Некорректно указано отчество"
	ErrInvalidLastName     = "Некорректно указана фамилия"
	ErrInvalidEmail        = "Некорректный адрес электронной почты"
	ErrInvalidBirthDate    = "Некорректная дата рождения. Регистрация доступна для лиц, достигших 14 лет"
	ErrInvalidPassword     = "Пароль должен состоять из 8 и более символов и содержать как минимум одну латинскую букву и одну цифру"
	ErrInvalidRefreshToken = "Не указан refresh-токен"
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

	ErrSamePassword = &APIError{
		Code:    "same_password",
		Message: "Новый пароль не должен совпадать со старым",
		Status:  http.StatusBadRequest,
	}
	ErrInvalidRequest = &APIError{
		Code:    "invalid_request",
		Message: "Некорректный запрос",
		Status:  http.StatusBadRequest,
	}

	ErrNotFound = &APIError{
		Code:    "not_found",
		Message: "Запрошенный ресурс не найден",
		Status:  http.StatusNotFound,
	}

	ErrForbidden = &APIError{
		Code:    "forbidden",
		Message: "Недостаточно прав для выполнения операции",
		Status:  http.StatusForbidden,
	}

	ErrInvalidDocument = &APIError{
		Code:    "invalid_document",
		Message: "Некорректно указаны данные документа",
		Status:  http.StatusUnprocessableEntity,
	}

	ErrDocumentNotAllowed = &APIError{
		Code:    "document_not_allowed",
		Message: "Указанный документ не подходит для выбранного маршрута",
		Status:  http.StatusUnprocessableEntity,
	}

	ErrDocumentRejected = &APIError{
		Code:    "document_rejected",
		Message: "Документ не прошёл проверку",
		Status:  http.StatusUnprocessableEntity,
	}

	ErrDocumentExpired = &APIError{
		Code:    "document_expired",
		Message: "Срок действия документа истёк к дате поездки",
		Status:  http.StatusUnprocessableEntity,
	}

	ErrDocumentAlreadyExists = &APIError{
		Code:    "document_already_exists",
		Message: "Такой документ уже сохранён",
		Status:  http.StatusConflict,
	}

	ErrPassengerCountMismatch = &APIError{
		Code:    "passenger_count_mismatch",
		Message: "Количество пассажиров не совпадает с параметрами поиска",
		Status:  http.StatusConflict,
	}

	ErrOrderExpired = &APIError{
		Code:    "order_expired",
		Message: "Срок оплаты заказа истёк",
		Status:  http.StatusConflict,
	}

	ErrInsufficientBonus = &APIError{
		Code:    "insufficient_bonus",
		Message: "Недостаточно бонусов для оплаты заказа",
		Status:  http.StatusConflict,
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
