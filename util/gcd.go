package util

import (
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
)

// calculate two numbers Greatest common divisor.
func GCD2(a, b int) int {
	for a != b {
		if a > b {
			a -= b
		} else {
			b -= a
		}
	}

	return a
}

// caculate the Greatest common divisor of the array.
func GCD(array []int) int {
	if len(array) == 0 {
		client_logger.GetMindAlphaServingClientLogger().Errorf("GCD(): len(array) = 0")
		return -1
	}
	if len(array) == 1 {
		return array[0]
	}
	if len(array) == 2 {
		return GCD2(array[0], array[1])
	}

	var g int = 0
	for _, v := range array {
		if v > 0 {
			if g > 0 {
				g = GCD2(v, g)
			} else {
				g = v
			}
		}
	}
	if 0 == g {
		return 1
	} else {
		return g
	}

}
