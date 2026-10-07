package models

type City struct {
	PostalCode string        `json:"postal_code" gorm:"primaryKey;column:postal_code"`
	Name       string        `json:"name" gorm:"column:name;not null"`
	Terminals  []BusTerminal `json:"terminals,omitempty" gorm:"foreignKey:PostalCode;references:PostalCode"`
}

func (City) TableName() string {
	return "city"
}

type CreateCityRequest struct {
	PostalCode string `json:"postal_code"`
	Name       string `json:"name"`
}

type UpdateCityRequest struct {
	PostalCode *string `json:"postal_code,omitempty"`
	Name       *string `json:"name,omitempty"`
}

// ListCitiesParams son los parámetros de paginación de GET /api/admin/cities.
type ListCitiesParams struct {
	Page  int    // default 1
	Limit int    // default 10
	Order string // "ASC" o "DESC", default "DESC"
}

// ListCitiesResponse es el payload paginado de ciudades.
type ListCitiesResponse struct {
	Cities        []City `json:"cities"`
	Page          int    `json:"page"`
	Next          *int   `json:"next"`
	Prev          *int   `json:"prev"`
	Elements      int    `json:"elements"`
	TotalElements int64  `json:"total_elements"`
}

// CountCitiesResponse es el payload de GET /api/admin/cities/count.
type CountCitiesResponse struct {
	Total int64 `json:"total"`
}

// CountActiveResponse es el payload de los endpoints GET /api/admin/stats/*/active.
type CountActiveResponse struct {
	Total int64 `json:"total"`
}
