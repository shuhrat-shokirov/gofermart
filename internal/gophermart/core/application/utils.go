package application

const (
	centsInPound = 100
)

func convertToPence(amount float64) int {
	return int(amount * centsInPound)
}

func convertToPounds(amount int) float64 {
	return float64(amount) / centsInPound
}
