package utils

import (
	"fmt"
	"time"
)


func ValidateRequired(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s wajib diisi", fieldName)
	}
	return nil
}


func ValidateWeight(weight float64) error {
	if weight < 0 || weight > 2000 {
		return fmt.Errorf("berat harus antara 0 - 2000 Kg, diterima: %.4f", weight)
	}
	return nil
}


func ValidateGender(gender string) error {
	if gender != "jantan" && gender != "betina" {
		return fmt.Errorf("gender harus 'jantan' atau 'betina', diterima: %s", gender)
	}
	return nil
}


func ValidateStatus(status string) error {
	if status != "active" && status != "sold" && status != "deceased" {
		return fmt.Errorf("status harus 'active', 'sold', atau 'deceased', diterima: %s", status)
	}
	return nil
}


func ValidateDateNotFuture(dateStr string) error {
	if dateStr == "" {
		return nil 
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("format tanggal harus YYYY-MM-DD, diterima: %s", dateStr)
	}
	if date.After(time.Now()) {
		return fmt.Errorf("tanggal tidak boleh di masa depan: %s", dateStr)
	}
	return nil
}


func ValidateRole(role string) error {
	if role != "admin" && role != "peternak" {
		return fmt.Errorf("role harus 'admin' atau 'peternak', diterima: %s", role)
	}
	return nil
}
