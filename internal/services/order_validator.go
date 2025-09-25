package services

type OrderValidator interface {
	Validate(order string) bool
}

type LunhOrderValidator struct {
}

func (v *LunhOrderValidator) Validate(number string) bool {
	sum := 0
	isSecond := false

	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')

		if isSecond {
			d = d * 2
		}

		sum += d / 10
		sum += d % 10

		isSecond = !isSecond
	}

	return (sum % 10) == 0
}
