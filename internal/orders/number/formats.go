package number

func randomAirPrefix() (string, error) {
	letters, err := randomLetters(airPrefixLetters)
	if err != nil {
		return "", err
	}
	digits, err := randomDigits(airPrefixDigits)
	if err != nil {
		return "", err
	}
	return letters + "-" + digits, nil
}

func randomRailPrefix() (string, error) {
	first, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	letter, err := randomLetters(1)
	if err != nil {
		return "", err
	}
	last, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	return first + letter + last, nil
}

func randomRailTail() (string, error) {
	digit, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	firstLetter, err := randomLetters(1)
	if err != nil {
		return "", err
	}
	secondDigit, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	lastLetter, err := randomLetters(1)
	if err != nil {
		return "", err
	}
	return digit + firstLetter + secondDigit + lastLetter, nil
}

func randomBusSuffix() (string, error) {
	digit, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	letter, err := randomLetters(1)
	if err != nil {
		return "", err
	}
	return digit + letter, nil
}
