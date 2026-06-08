package catalog

import (
	"net/http"

	"fixapp/internal/domain"
	"fixapp/pkg/response"

	"go.uber.org/zap"
)

// Handler handles HTTP requests for catalog (reference data) operations.
type Handler struct {
	repo   Repository
	logger *zap.Logger
}

// NewHandler creates a new catalog handler.
func NewHandler(repo Repository, logger *zap.Logger) *Handler {
	return &Handler{
		repo:   repo,
		logger: logger,
	}
}

// Register registers the catalog routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /categories", h.ListCategories)
	mux.HandleFunc("GET /districts", h.ListDistricts)
}

// CategoryResponse is the API representation of a service category.
// @Description Service category
type CategoryResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string `json:"name" example:"Hydraulik"`
	Slug      string `json:"slug" example:"hydraulik"`
	Icon      string `json:"icon,omitempty" example:"wrench"`
	BasePrice int    `json:"base_price" example:"28"`
}

// DistrictResponse is the API representation of a district.
// @Description District/area
type DistrictResponse struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name     string `json:"name" example:"Krowodrza"`
	Slug     string `json:"slug" example:"krowodrza"`
	CityName string `json:"city_name" example:"Krakow"`
}

// CategoriesListResponse wraps the categories list.
// @Description List of service categories
type CategoriesListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

// DistrictsListResponse wraps the districts list.
// @Description List of districts
type DistrictsListResponse struct {
	Districts []DistrictResponse `json:"districts"`
}

// ListCategories godoc
// @Summary      List service categories
// @Description  Returns all active service categories
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Success      200  {object}  CategoriesListResponse
// @Router       /categories [get]
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.ListCategories(r.Context(), true)
	if err != nil {
		h.logger.Error("failed to list categories", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	resp := CategoriesListResponse{
		Categories: toCategoryResponses(categories),
	}
	response.OK(w, resp)
}

// ListDistricts godoc
// @Summary      List districts
// @Description  Returns all active districts
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Success      200  {object}  DistrictsListResponse
// @Router       /districts [get]
func (h *Handler) ListDistricts(w http.ResponseWriter, r *http.Request) {
	districts, err := h.repo.ListDistricts(r.Context(), true)
	if err != nil {
		h.logger.Error("failed to list districts", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	resp := DistrictsListResponse{
		Districts: toDistrictResponses(districts),
	}
	response.OK(w, resp)
}

func toCategoryResponses(categories []*domain.ServiceCategory) []CategoryResponse {
	resp := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		resp[i] = CategoryResponse{
			ID:        c.ID.String(),
			Name:      c.Name,
			Slug:      c.Slug,
			Icon:      c.Icon,
			BasePrice: c.BasePrice,
		}
	}
	return resp
}

func toDistrictResponses(districts []*domain.District) []DistrictResponse {
	resp := make([]DistrictResponse, len(districts))
	for i, d := range districts {
		resp[i] = DistrictResponse{
			ID:       d.ID.String(),
			Name:     d.Name,
			Slug:     d.Slug,
			CityName: d.CityName,
		}
	}
	return resp
}
