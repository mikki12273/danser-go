package mutils

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
	"strings"
)

// ClampF is Clamp but optimized for floats
func ClampF[T constraints.Float](x, min, max T) T {
	return T(math.Min(float64(max), math.Max(float64(min), float64(x))))
}

func Abs[T constraints.Integer | constraints.Float](a T) T {
	if a < 0 {
		return -a
	}

	return a
}

func Min[T constraints.Integer | constraints.Float](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func Max[T constraints.Integer | constraints.Float](a, b T) T {
	if a > b {
		return a
	}

	return b
}

func Clamp[T constraints.Integer | constraints.Float](x, min, max T) T {
	return Min(max, Max(min, x))
}

func Lerp[T constraints.Integer | constraints.Float, V constraints.Float](min, max T, t V) T {
	return min + T(V(max-min)*t)
}

func Compare[T constraints.Integer | constraints.Float](a, b T) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}

	return 0
}

func Signum[T constraints.Float](a T) T {
	if a == 0 {
		return 0
	}

	if math.Signbit(float64(a)) {
		return -1
	}

	return 1
}

func SanitizeAngle[T constraints.Float](a T) T {
	a = T(math.Mod(float64(a), 2*math.Pi))
	if a < 0 {
		a += T(2 * math.Pi)
	}

	return a
}

func SanitizeAngleArc[T constraints.Float](a T) T {
	sPi := T(math.Pi)

	if a < -sPi {
		a += 2 * sPi
	} else if a >= sPi {
		a -= 2 * sPi
	}

	return a
}

// FormatWOZeros formats the float with specified precision but removes trailing zeros
func FormatWOZeros[T constraints.Float](val T, precision int) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.*f", precision, val), "0"), ".")
}
