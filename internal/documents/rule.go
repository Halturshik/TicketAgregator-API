package documents

type DocumentRule struct {
	Transport                  string
	IsInternational            bool
	MinAge                     int
	MaxAge                     int
	IsRussian                  bool
	AllowInternalPassport      bool
	AllowInternationalPassport bool
	AllowBirthCertificate      bool
	AllowForeignPassport       bool
}

func (r DocumentRule) Allows(documentType string) bool {
	switch documentType {
	case TypeInternalPassport:
		return r.AllowInternalPassport
	case TypeInternationalPassport:
		return r.AllowInternationalPassport
	case TypeBirthCertificate:
		return r.AllowBirthCertificate
	case TypeForeignPassport:
		return r.AllowForeignPassport
	default:
		return false
	}
}
