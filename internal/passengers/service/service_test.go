package service

import "testing"

func TestValidPassengerBirthDateAllowsChildren(t *testing.T) {
	if _, err := validPassengerBirthDate("2022-01-02"); err != nil {
		t.Fatalf("expected child passenger birth date to be valid: %v", err)
	}
}

func TestValidPassengerBirthDateRejectsBadFormat(t *testing.T) {
	if _, err := validPassengerBirthDate("02.01.2022"); err == nil {
		t.Fatalf("expected non-ISO birth date to be invalid")
	}
}
