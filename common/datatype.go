package common

import "fmt"

type Fraction struct {
	// The numerator in the fraction, e.g. 2 in 2/3.
	Numerator int32
	// The value by which the numerator is divided, e.g. 3 in 2/3.
	Denominator int32
}

func (receiver Fraction) String() string {
	if receiver.Numerator == 0 {
		return "0"
	} else {
		return fmt.Sprintf("%d/%d", receiver.Numerator, receiver.Denominator)
	}
}

func (receiver Fraction) ExposureCompensationString() string {
	if receiver.Numerator == 0 {
		return "0"
	} else if receiver.Numerator/receiver.Denominator > 0 {
		return fmt.Sprintf("+%d/%d", receiver.Numerator, receiver.Denominator)
	} else {
		return fmt.Sprintf("%d/%d", receiver.Numerator, receiver.Denominator)
	}
}

type UFraction struct {
	// The numerator in the fraction, e.g. 2 in 2/3.
	Numerator uint32
	// The value by which the numerator is divided, e.g. 3 in 2/3.
	Denominator uint32
}

func (receiver UFraction) String() string {
	return fmt.Sprintf("%d/%d", receiver.Numerator, receiver.Denominator)
}

func (receiver UFraction) ValString() string {
	return fmt.Sprintf("%.2f", float32(receiver.Numerator)/float32(receiver.Denominator))
}

func (receiver UFraction) FocalLengthFormat() string {
	return fmt.Sprintf("%.0f", float32(receiver.Numerator)/float32(receiver.Denominator))
}

func (receiver UFraction) FNumberFormat() string {
	if receiver.Numerator == 0 {
		return "0"
	}
	val := float32(receiver.Numerator) / float32(receiver.Denominator)
	if val < 1 {
		return fmt.Sprintf("%.2f", val)
	} else {
		return fmt.Sprintf("%.1f", val)
	}
}

func (receiver UFraction) ShutterFormat() string {
	val := float32(receiver.Numerator) / float32(receiver.Denominator)
	if val <= 0.25 && val > 0 {
		return fmt.Sprintf("1/%.0f", float32(receiver.Denominator)/float32(receiver.Numerator))
	} else {
		return fmt.Sprintf("%.1f", float32(receiver.Numerator)/float32(receiver.Denominator))
	}
}
