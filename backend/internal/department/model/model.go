package model

import "time"

type Department struct {
 PublicID string `json:"id"`
 Name string `json:"name"`
 Code string `json:"code,omitempty"`
 Description string `json:"description,omitempty"`
 IsActive bool `json:"is_active"`
 EmployeeCount int `json:"employee_count"`
 CreatedAt time.Time `json:"created_at"`
}

type CreateInput struct { Name string `json:"name"`; Code string `json:"code"`; Description string `json:"description"` }
type UpdateInput struct { Name string `json:"name"`; Code string `json:"code"`; Description string `json:"description"`; IsActive bool `json:"is_active"` }