package main

const hourInSeconds = 3600
const unSetSeconds = 0

func setExperation(seconds int) int {
	if seconds <= unSetSeconds || seconds >= hourInSeconds {
		return hourInSeconds
	}
	return seconds
}
