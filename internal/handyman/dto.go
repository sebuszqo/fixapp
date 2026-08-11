package handyman

import (
	"time"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// ===== Response DTOs =====

// ===== Response DTOs =====

// ProfileResponse is the public representation of a handyman profile.
// @Description Handyman profile information
type ProfileResponse struct {
	ID                          string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID                      string    `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	CompanyName                 string    `json:"company_name,omitempty" example:"Jan Kowalski Hydraulika"`
	NIP                         string    `json:"nip,omitempty" example:"1234567890"`
	CompanyAddress              string    `json:"company_address,omitempty" example:"Kraków, ul. Przykładowa 1"`
	PKDCode                     string    `json:"pkd_code,omitempty" example:"4322Z - Wykonywanie instalacji wod-kan"`
	GUSStatus                   string    `json:"gus_status,omitempty" example:"Aktywna"`
	GUSVerified                 bool      `json:"gus_verified"`
	Phone                       string    `json:"phone,omitempty" example:"+48123456789"`
	Email                       string    `json:"email,omitempty" example:"jan@example.com"`
	BusinessType                string    `json:"business_type,omitempty" example:"jdg"`
	IsVATPayer                  bool      `json:"is_vat_payer"`
	ConsentIdentityVerification bool      `json:"consent_identity_verification"`
	ConsentMarketing            bool      `json:"consent_marketing"`
	Bio                         string    `json:"bio,omitempty" example:"Hydraulik z 10-letnim doswiadczeniem"`
	AvatarURL                   string    `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	ExperienceYears             string    `json:"experience_years,omitempty" example:"6-10 lat"`
	WorkingHours                string    `json:"working_hours,omitempty"`
	Categories                  []string  `json:"categories"`
	Districts                   []string  `json:"districts"`
	IsAvailable                 bool      `json:"is_available" example:"true"`
	EmergencyAvailable          bool      `json:"emergency_available" example:"false"`
	IsVerified                  bool      `json:"is_verified" example:"true"`
	VerificationStatus          string    `json:"verification_status" example:"unverified"`
	CEIDGDocumentURL            string    `json:"ceidg_document_url,omitempty"`
	MicrotransferStatus         string    `json:"microtransfer_status,omitempty" example:"none"`
	MicrotransferCode           string    `json:"microtransfer_code,omitempty" example:"FIXAPP-7823-JAN"`
	CompletionPct               int       `json:"completion_pct" example:"85"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

// ProfileListResponse is the paginated list of profiles.
// @Description Paginated handyman profile list
type ProfileListResponse struct {
	Profiles []ProfileResponse `json:"profiles"`
	Total    int64             `json:"total" example:"50"`
	Limit    int               `json:"limit" example:"20"`
	Offset   int               `json:"offset" example:"0"`
	HasMore  bool              `json:"has_more" example:"true"`
}

// PricingItemResponse represents a single pricing item.
// @Description Pricing list item
type PricingItemResponse struct {
	ID                string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ServiceName       string `json:"service_name" example:"Wymiana kranu"`
	PriceFrom         int    `json:"price_from" example:"100"`
	PriceTo           *int   `json:"price_to,omitempty" example:"200"`
	Unit              string `json:"unit" example:"per service"`
	EstimatedDuration string `json:"estimated_duration,omitempty" example:"~1h"`
	SortOrder         int    `json:"sort_order" example:"0"`
}

// PortfolioItemResponse represents a portfolio photo.
// @Description Portfolio item
type PortfolioItemResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ImageURL  string    `json:"image_url" example:"https://example.com/photo.jpg"`
	Caption   string    `json:"caption,omitempty" example:"Remont lazienki"`
	SortOrder int       `json:"sort_order" example:"0"`
	CreatedAt time.Time `json:"created_at"`
}

// FullProfileResponse includes profile + pricing + portfolio.
// @Description Full handyman profile with pricing and portfolio
type FullProfileResponse struct {
	ProfileResponse
	Pricing   []PricingItemResponse   `json:"pricing"`
	Portfolio []PortfolioItemResponse `json:"portfolio"`
}

// GUSLookupResponse represents company info returned from GUS lookup.
type GUSLookupResponse struct {
	Found       bool   `json:"found"`
	NIP         string `json:"nip"`
	CompanyName string `json:"company_name"`
	Address     string `json:"address"`
	PKD         string `json:"pkd"`
	Status      string `json:"status"` // e.g. "Aktywna", "Zawieszona", "Nieznana"
	Message     string `json:"message,omitempty"`
}

// CEIDGUploadRequest is the payload for uploading CEIDG printout PDF.
type CEIDGUploadRequest struct {
	DocumentURL string `json:"document_url" example:"data:application/pdf;base64,..."`
}

// MicrotransferConfirmRequest is the payload for confirming microtransfer.
type MicrotransferConfirmRequest struct {
	Code string `json:"code" example:"FIXAPP-7823-JAN"`
}

// ===== Request DTOs =====

// UpdateProfileRequest is the payload for updating a handyman profile.
// @Description Update handyman profile
type UpdateProfileRequest struct {
	CompanyName                 *string  `json:"company_name,omitempty" example:"Jan Kowalski Hydraulika"`
	NIP                         *string  `json:"nip,omitempty" example:"1234567890"`
	CompanyAddress              *string  `json:"company_address,omitempty"`
	PKDCode                     *string  `json:"pkd_code,omitempty"`
	GUSStatus                   *string  `json:"gus_status,omitempty"`
	GUSVerified                 *bool    `json:"gus_verified,omitempty"`
	Phone                       *string  `json:"phone,omitempty" example:"+48123456789"`
	Email                       *string  `json:"email,omitempty" example:"jan@example.com"`
	BusinessType                *string  `json:"business_type,omitempty"`
	IsVATPayer                  *bool    `json:"is_vat_payer,omitempty"`
	ConsentIdentityVerification *bool    `json:"consent_identity_verification,omitempty"`
	ConsentMarketing            *bool    `json:"consent_marketing,omitempty"`
	Bio                         *string  `json:"bio,omitempty" example:"Hydraulik z 10-letnim doswiadczeniem"`
	AvatarURL                   *string  `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	ExperienceYears             *string  `json:"experience_years,omitempty"`
	WorkingHours                *string  `json:"working_hours,omitempty"`
	Categories                  []string `json:"categories,omitempty"`
	Districts                   []string `json:"districts,omitempty"`
	IsAvailable                 *bool    `json:"is_available,omitempty" example:"true"`
	EmergencyAvailable          *bool    `json:"emergency_available,omitempty" example:"false"`
	VerificationStatus          *string  `json:"verification_status,omitempty"`
	CEIDGDocumentURL            *string  `json:"ceidg_document_url,omitempty"`
	MicrotransferStatus         *string  `json:"microtransfer_status,omitempty"`
	MicrotransferCode           *string  `json:"microtransfer_code,omitempty"`
}

// Validate checks if the request is valid.
func (r *UpdateProfileRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Categories != nil && len(r.Categories) > 3 {
		errs["categories"] = "maximum 3 categories allowed"
	}
	if r.NIP != nil && *r.NIP != "" && len(*r.NIP) != 10 {
		errs["nip"] = "NIP must be exactly 10 digits"
	}
	if r.Bio != nil && len(*r.Bio) > 2000 {
		errs["bio"] = "bio must be 2000 characters or less"
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// CreatePricingRequest is the payload for adding a pricing item.
// @Description Create pricing item
type CreatePricingRequest struct {
	ServiceName       string `json:"service_name" example:"Wymiana kranu"`
	PriceFrom         int    `json:"price_from" example:"100"`
	PriceTo           *int   `json:"price_to,omitempty" example:"200"`
	Unit              string `json:"unit,omitempty" example:"per service"`
	EstimatedDuration string `json:"estimated_duration,omitempty" example:"~1h"`
	SortOrder         int    `json:"sort_order,omitempty" example:"0"`
}

// Validate checks if the request is valid.
func (r *CreatePricingRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.ServiceName == "" {
		errs["service_name"] = "service_name is required"
	}
	if r.PriceFrom < 0 {
		errs["price_from"] = "price_from must be non-negative"
	}
	if r.PriceTo != nil && *r.PriceTo < r.PriceFrom {
		errs["price_to"] = "price_to must be >= price_from"
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// AddPortfolioRequest is the payload for adding a portfolio photo.
// @Description Add portfolio item
type AddPortfolioRequest struct {
	ImageURL  string `json:"image_url" example:"https://example.com/photo.jpg"`
	Caption   string `json:"caption,omitempty" example:"Remont lazienki"`
	SortOrder int    `json:"sort_order,omitempty" example:"0"`
}

// Validate checks if the request is valid.
func (r *AddPortfolioRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.ImageURL == "" {
		errs["image_url"] = "image_url is required"
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ===== Mappers =====

func uuidSliceToStrings(ids []uuid.UUID) []string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.String()
	}
	return strs
}

// ToProfileResponse converts a domain profile to API response.
func ToProfileResponse(p *domain.HandymanProfile) ProfileResponse {
	return ProfileResponse{
		ID:                          p.ID.String(),
		UserID:                      p.UserID.String(),
		CompanyName:                 p.CompanyName,
		NIP:                         p.NIP,
		CompanyAddress:              p.CompanyAddress,
		PKDCode:                     p.PKDCode,
		GUSStatus:                   p.GUSStatus,
		GUSVerified:                 p.GUSVerified,
		Phone:                       p.Phone,
		Email:                       p.Email,
		BusinessType:                p.BusinessType,
		IsVATPayer:                  p.IsVATPayer,
		ConsentIdentityVerification: p.ConsentIdentityVerification,
		ConsentMarketing:            p.ConsentMarketing,
		Bio:                         p.Bio,
		AvatarURL:                   p.AvatarURL,
		ExperienceYears:             p.ExperienceYears,
		WorkingHours:                p.WorkingHours,
		Categories:                  uuidSliceToStrings(p.Categories),
		Districts:                   uuidSliceToStrings(p.Districts),
		IsAvailable:                 p.IsAvailable,
		EmergencyAvailable:          p.EmergencyAvailable,
		IsVerified:                  p.IsVerified,
		VerificationStatus:          p.VerificationStatus,
		CEIDGDocumentURL:            p.CEIDGDocumentURL,
		MicrotransferStatus:         p.MicrotransferStatus,
		MicrotransferCode:           p.MicrotransferCode,
		CompletionPct:               p.CompletionPercentage(),
		CreatedAt:                   p.CreatedAt,
		UpdatedAt:                   p.UpdatedAt,
	}
}

// ToProfileListResponse converts a list of profiles to paginated response.
func ToProfileListResponse(profiles []*domain.HandymanProfile, total int64, limit, offset int) ProfileListResponse {
	responses := make([]ProfileResponse, len(profiles))
	for i, p := range profiles {
		responses[i] = ToProfileResponse(p)
	}
	return ProfileListResponse{
		Profiles: responses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		HasMore:  int64(offset+len(profiles)) < total,
	}
}

// ToPricingItemResponse converts a domain pricing item to API response.
func ToPricingItemResponse(item *domain.PricingItem) PricingItemResponse {
	return PricingItemResponse{
		ID:                item.ID.String(),
		ServiceName:       item.ServiceName,
		PriceFrom:         item.PriceFrom,
		PriceTo:           item.PriceTo,
		Unit:              item.Unit,
		EstimatedDuration: item.EstimatedDuration,
		SortOrder:         item.SortOrder,
	}
}

// ToPortfolioItemResponse converts a domain portfolio item to API response.
func ToPortfolioItemResponse(item *domain.PortfolioItem) PortfolioItemResponse {
	return PortfolioItemResponse{
		ID:        item.ID.String(),
		ImageURL:  item.ImageURL,
		Caption:   item.Caption,
		SortOrder: item.SortOrder,
		CreatedAt: item.CreatedAt,
	}
}

// ToFullProfileResponse assembles a full profile response.
func ToFullProfileResponse(p *domain.HandymanProfile, pricing []*domain.PricingItem, portfolio []*domain.PortfolioItem) FullProfileResponse {
	pricingResp := make([]PricingItemResponse, len(pricing))
	for i, item := range pricing {
		pricingResp[i] = ToPricingItemResponse(item)
	}

	portfolioResp := make([]PortfolioItemResponse, len(portfolio))
	for i, item := range portfolio {
		portfolioResp[i] = ToPortfolioItemResponse(item)
	}

	return FullProfileResponse{
		ProfileResponse: ToProfileResponse(p),
		Pricing:         pricingResp,
		Portfolio:        portfolioResp,
	}
}

