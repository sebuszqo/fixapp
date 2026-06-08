package domain

import (
	"github.com/google/uuid"
)

// ServiceCategory represents a type of service (e.g., Hydraulik, Elektryk).
type ServiceCategory struct {
	ID       uuid.UUID
	Name     string // e.g., "Hydraulik"
	Slug     string // e.g., "hydraulik" (URL-friendly)
	Icon     string // icon identifier for frontend
	BasePrice int    // base lead fee in credits for this category
	IsActive bool
}

// District represents a geographic area within the city.
type District struct {
	ID       uuid.UUID
	Name     string // e.g., "Krowodrza"
	Slug     string // e.g., "krowodrza"
	CityName string // e.g., "Kraków"
	IsActive bool
}
