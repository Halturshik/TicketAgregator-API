package store

const keyPrefix = "booking_access:"

func challengeKey(id string) string {
	return keyPrefix + "challenge:" + id
}

func attemptsKey(id string) string {
	return keyPrefix + "attempts:" + id
}

func requestKey(identity string) string {
	return keyPrefix + "request:" + identity
}

func cooldownKey(identity string) string {
	return keyPrefix + "cooldown:" + identity
}

func activeChallengeKey(identity string) string {
	return keyPrefix + "active:" + identity
}

func verifyLockKey(id string) string {
	return keyPrefix + "lock:verify:" + id
}

func tokenKey(hash string) string {
	return keyPrefix + "token:" + hash
}
