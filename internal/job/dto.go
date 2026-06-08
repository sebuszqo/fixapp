package job

import (
	"time"

	"fixapp/internal/domain"
)

// ===== Response DTOs =====

// JobResponse is the public representation of a job.
// @Description Job information
type JobResponse struct {
	ID          string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ClientID    string     `json:"client_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	CategoryID  string     `json:"category_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	DistrictID  string     `json:"district_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	Title       string     `json:"title" example:"Cieknacy kran w kuchni"`
	Description string     `json:"description" example:"Kran cieknie od tygodnia, potrzebna wymiana uszczelki"`
	Urgency     string     `json:"urgency" example:"normal"`
	Status      string     `json:"status" example:"active"`
	Address     string     `json:"address,omitempty" example:"ul. Krolewska 10/5"`
	BuildingType string    `json:"building_type,omitempty" example:"apartment"`
	Floor       *int       `json:"floor,omitempty" example:"3"`
	HasElevator *bool      `json:"has_elevator,omitempty" example:"true"`
	PreferredDate1 *time.Time `json:"preferred_date1,omitempty"`
	PreferredDate2 *time.Time `json:"preferred_date2,omitempty"`
	PreferredTime  string     `json:"preferred_time,omitempty" example:"morning"`
	Budget         *int       `json:"budget,omitempty" example:"200"`
	WantsInvoice   bool       `json:"wants_invoice" example:"false"`
	ContactMethod  string     `json:"contact_method" example:"phone"`
	PhotoURLs      []string   `json:"photo_urls"`
	FinalValue     *int       `json:"final_value,omitempty" example:"280"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CompletedByID  *string    `json:"completed_by_id,omitempty"`
	ClientConfirmed *bool     `json:"client_confirmed,omitempty"`
	CreatedAt      time.Time  `json:"created_at" example:"2025-01-10T12:00:00Z"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2025-01-10T12:00:00Z"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// JobListResponse is the paginated list of jobs.
// @Description Paginated job list
type JobListResponse struct {
	Jobs    []JobResponse `json:"jobs"`
	Total   int64         `json:"total" example:"50"`
	Limit   int           `json:"limit" example:"20"`
	Offset  int           `json:"offset" example:"0"`
	HasMore bool          `json:"has_more" example:"true"`
}

// ===== Request DTOs =====

// CreateJobRequest is the payload for creating a new job.
// @Description Create job request
type CreateJobRequest struct {
	CategoryID     string `json:"category_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	DistrictID     string `json:"district_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	Title          string `json:"title" example:"Cieknacy kran w kuchni"`
	Description    string `json:"description" example:"Kran cieknie od tygodnia"`
	Urgency        string `json:"urgency,omitempty" example:"normal"`
	Address        string `json:"address,omitempty" example:"ul. Krolewska 10/5"`
	BuildingType   string `json:"building_type,omitempty" example:"apartment"`
	Floor          *int   `json:"floor,omitempty" example:"3"`
	HasElevator    *bool  `json:"has_elevator,omitempty" example:"true"`
	PreferredDate1 *time.Time `json:"preferred_date1,omitempty"`
	PreferredDate2 *time.Time `json:"preferred_date2,omitempty"`
	PreferredTime  string     `json:"preferred_time,omitempty" example:"morning"`
	Budget         *int       `json:"budget,omitempty" example:"200"`
	WantsInvoice   bool       `json:"wants_invoice,omitempty" example:"false"`
	ContactMethod  string     `json:"contact_method,omitempty" example:"phone"`
	PhotoURLs      []string   `json:"photo_urls,omitempty"`
}

// Validate checks if the request is valid.
func (r *CreateJobRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.CategoryID == "" {
		errs["category_id"] = "category_id is required"
	}
	if r.DistrictID == "" {
		errs["district_id"] = "district_id is required"
	}
	if r.Title == "" {
		errs["title"] = "title is required"
	} else if len(r.Title) > 200 {
		errs["title"] = "title must be 200 characters or less"
	}
	if r.Description == "" {
		errs["description"] = "description is required"
	}
	if r.Urgency != "" {
		u := domain.JobUrgency(r.Urgency)
		if !u.IsValid() {
			errs["urgency"] = "must be one of: low, normal, urgent, emergency"
		}
	}
	if r.PreferredTime != "" {
		valid := map[string]bool{"morning": true, "afternoon": true, "evening": true, "flexible": true}
		if !valid[r.PreferredTime] {
			errs["preferred_time"] = "must be one of: morning, afternoon, evening, flexible"
		}
	}
	if r.ContactMethod != "" {
		valid := map[string]bool{"phone": true, "app": true, "any": true}
		if !valid[r.ContactMethod] {
			errs["contact_method"] = "must be one of: phone, app, any"
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// CompleteJobRequest is the payload for marking a job as completed.
// @Description Complete job request
type CompleteJobRequest struct {
	FinalValue int `json:"final_value" example:"280"`
}

// Validate checks if the request is valid.
func (r *CompleteJobRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.FinalValue <= 0 {
		errs["final_value"] = "final_value must be a positive integer (PLN)"
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ===== Mappers =====

// ToJobResponse converts a domain job to API response.
func ToJobResponse(job *domain.Job) JobResponse {
	resp := JobResponse{
		ID:             job.ID.String(),
		ClientID:       job.ClientID.String(),
		CategoryID:     job.CategoryID.String(),
		DistrictID:     job.DistrictID.String(),
		Title:          job.Title,
		Description:    job.Description,
		Urgency:        job.Urgency.String(),
		Status:         job.Status.String(),
		Address:        job.Address,
		BuildingType:   job.BuildingType,
		Floor:          job.Floor,
		HasElevator:    job.HasElevator,
		PreferredDate1: job.PreferredDate1,
		PreferredDate2: job.PreferredDate2,
		PreferredTime:  job.PreferredTime,
		Budget:         job.Budget,
		WantsInvoice:   job.WantsInvoice,
		ContactMethod:  job.ContactMethod,
		PhotoURLs:      job.PhotoURLs,
		FinalValue:     job.FinalValue,
		CompletedAt:    job.CompletedAt,
		ClientConfirmed: job.ClientConfirmed,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
		ExpiresAt:      job.ExpiresAt,
	}
	if job.CompletedByID != nil {
		s := job.CompletedByID.String()
		resp.CompletedByID = &s
	}
	if resp.PhotoURLs == nil {
		resp.PhotoURLs = []string{}
	}
	return resp
}

// ToJobListResponse converts a list of domain jobs to paginated response.
func ToJobListResponse(jobs []*domain.Job, total int64, limit, offset int) JobListResponse {
	responses := make([]JobResponse, len(jobs))
	for i, j := range jobs {
		responses[i] = ToJobResponse(j)
	}
	return JobListResponse{
		Jobs:    responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(offset+len(jobs)) < total,
	}
}
