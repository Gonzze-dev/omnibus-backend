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
	Name string `json:"name"`
}
