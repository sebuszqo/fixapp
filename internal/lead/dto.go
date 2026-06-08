package lead

import (
	"time"

	"fixapp/internal/domain"
)

// ===== Response DTOs =====

// LeadResponse is the public representation of a lead.
// @Description Lead information
type LeadResponse struct {
	ID                string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	JobID             string     `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	HandymanID        string     `json:"handyman_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Status            string     `json:"status" example:"pending"`
	Price             int        `json:"price" example:"22"`
	ClientCommitScore int        `json:"client_commit_score" example:"85"`
	CreatedAt         time.Time  `json:"created_at" example:"2025-01-10T12:00:00Z"`
	UpdatedAt         time.Time  `json:"updated_at" example:"2025-01-10T12:00:00Z"`
	ExpiresAt         time.Time  `json:"expires_at" example:"2025-01-11T12:00:00Z"`
	AcceptedAt        *time.Time `json:"accepted_at,omitempty"`
	RejectedAt        *time.Time `json:"rejected_at,omitempty"`
}

// LeadDetailResponse includes the lead plus job details.
// @Description Lead with job details
type LeadDetailResponse struct {
	LeadResponse
	Job *LeadJobInfo `json:"job"`
}

// LeadJobInfo is a subset of job info shown to handymen on a lead.
// @Description Job info visible on a lead
type LeadJobInfo struct {
	Title          string     `json:"title" example:"Cieknacy kran w kuchni"`
	Description    string     `json:"description" example:"Kran cieknie od tygodnia"`
	CategoryID     string     `json:"category_id"`
	DistrictID     string     `json:"district_id"`
	Urgency        string     `json:"urgency" example:"normal"`
	BuildingType   string     `json:"building_type,omitempty" example:"apartment"`
	Floor          *int       `json:"floor,omitempty"`
	HasElevator    *bool      `json:"has_elevator,omitempty"`
	PreferredDate1 *time.Time `json:"preferred_date1,omitempty"`
	PreferredDate2 *time.Time `json:"preferred_date2,omitempty"`
	PreferredTime  string     `json:"preferred_time,omitempty" example:"morning"`
	Budget         *int       `json:"budget,omitempty"`
	PhotoURLs      []string   `json:"photo_urls"`
	ContactMethod  string     `json:"contact_method" example:"phone"`
	// Address is only revealed after acceptance
	Address string `json:"address,omitempty" example:"ul. Krolewska 10/5"`
}

// LeadListResponse is the paginated list of leads.
// @Description Paginated lead list
type LeadListResponse struct {
	Leads   []LeadResponse `json:"leads"`
	Total   int64          `json:"total" example:"15"`
	Limit   int            `json:"limit" example:"20"`
	Offset  int            `json:"offset" example:"0"`
	HasMore bool           `json:"has_more" example:"false"`
}

// ===== Mappers =====

// ToLeadResponse converts a domain lead to API response.
func ToLeadResponse(lead *domain.Lead) LeadResponse {
	return LeadResponse{
		ID:                lead.ID.String(),
		JobID:             lead.JobID.String(),
		HandymanID:        lead.HandymanID.String(),
		Status:            lead.Status.String(),
		Price:             lead.Price,
		ClientCommitScore: lead.ClientCommitScore,
		CreatedAt:         lead.CreatedAt,
		UpdatedAt:         lead.UpdatedAt,
		ExpiresAt:         lead.ExpiresAt,
		AcceptedAt:        lead.AcceptedAt,
		RejectedAt:        lead.RejectedAt,
	}
}

// ToLeadDetailResponse converts a lead and its job to a detail response.
func ToLeadDetailResponse(lead *domain.Lead, job *domain.Job, revealAddress bool) LeadDetailResponse {
	jobInfo := &LeadJobInfo{
		Title:          job.Title,
		Description:    job.Description,
		CategoryID:     job.CategoryID.String(),
		DistrictID:     job.DistrictID.String(),
		Urgency:        job.Urgency.String(),
		BuildingType:   job.BuildingType,
		Floor:          job.Floor,
		HasElevator:    job.HasElevator,
		PreferredDate1: job.PreferredDate1,
		PreferredDate2: job.PreferredDate2,
		PreferredTime:  job.PreferredTime,
		Budget:         job.Budget,
		PhotoURLs:      job.PhotoURLs,
		ContactMethod:  job.ContactMethod,
	}
	if jobInfo.PhotoURLs == nil {
		jobInfo.PhotoURLs = []string{}
	}
	// Only reveal address after lead is accepted
	if revealAddress {
		jobInfo.Address = job.Address
	}

	return LeadDetailResponse{
		LeadResponse: ToLeadResponse(lead),
		Job:          jobInfo,
	}
}

// ToLeadListResponse converts a list of domain leads to paginated response.
func ToLeadListResponse(leads []*domain.Lead, total int64, limit, offset int) LeadListResponse {
	responses := make([]LeadResponse, len(leads))
	for i, l := range leads {
		responses[i] = ToLeadResponse(l)
	}
	return LeadListResponse{
		Leads:   responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(offset+len(leads)) < total,
	}
}
