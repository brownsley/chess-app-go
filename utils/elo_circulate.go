package utils

import "math"

func CalculateElo(ratingA, ratingB int, scoreA float64) (newRatingA, newRatingB, changeA, changeB int) {
	const kFactor = 32.0

	expectedA := 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
	expectedB := 1.0 - expectedA

	scoreB := 1.0 - scoreA

	deltaA := int(math.Round(kFactor * (scoreA - expectedA)))
	deltaB := int(math.Round(kFactor * (scoreB - expectedB)))

	newRatingA = ratingA + deltaA
	newRatingB = ratingB + deltaB

	return newRatingA, newRatingB, deltaA, deltaB
}
