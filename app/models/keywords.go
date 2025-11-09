package models

// BaseModel contains common fields for all models
type BaseModel struct {
	ID       int  `json:"id"`
	IsActive bool `json:"is_active"`
}

// Keyword represents a keyword in the database
type Keyword struct {
	BaseModel
	CategoryID int    `json:"category_id"`
	Keyword    string `json:"keyword"`
}

// Category represents a category in the database
type Category struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateStatus is used for updating the status of a keyword or category
type UpdateStatus struct {
	IsActive bool `json:"is_active"`
}
