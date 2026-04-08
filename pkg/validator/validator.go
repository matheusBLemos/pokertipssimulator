package validator

import "fmt"

func Required(field, name string) error {
	if field == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func MinMax(val, min, max int, name string) error {
	if val < min || val > max {
		return fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return nil
}
