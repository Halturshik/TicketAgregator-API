package payments

const (
	StatusFailed  = "failed"
	StatusSuccess = "success"
	ProviderMock  = "mock"
)

func ResultStatus(success bool) string {
	if success {
		return StatusSuccess
	}
	return StatusFailed
}
