package number

func NewOrderNumber() (string, error) {
	prefix, err := randomLetters(orderPrefixLetters)
	if err != nil {
		return "", err
	}
	suffix, err := randomDigits(orderSuffixDigits)
	if err != nil {
		return "", err
	}
	return prefix + "-" + suffix, nil
}
