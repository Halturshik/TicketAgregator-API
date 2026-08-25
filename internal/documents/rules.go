package documents

const (
	TypeInternalPassport      = "internal_passport"
	TypeInternationalPassport = "international_passport"
	TypeBirthCertificate      = "birth_certificate"
	TypeForeignPassport       = "foreign_passport"

	StatusVerified = "verified"
	StatusRejected = "rejected"

	DateLayout               = "2006-01-02"
	VerificationPeriodLayout = "2006-01"
	VerificationScale        = 100
	VerificationRejectRate   = 5
)

func IsAutoVerified(documentType string) bool {
	return documentType == TypeBirthCertificate || documentType == TypeForeignPassport
}
