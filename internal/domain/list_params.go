package domain

// ListAdsParams contains parameters for listing ads with pagination and filtering.
type ListAdsParams struct {
	// Sorting
	SortBy string // "price" or "created_at"
	Order  string // "asc" or "desc"
	
	// Pagination
	Page  int // 1-based page number
	Limit int // number of items per page
	
	// Filtering
	MinPrice *int64 // minimum price filter (optional)
	MaxPrice *int64 // maximum price filter (optional)
}

// GetOffset calculates the SQL OFFSET value from page and limit.
func (p *ListAdsParams) GetOffset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}

// SetDefaults sets default values for unspecified parameters.
func (p *ListAdsParams) SetDefaults() {
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.Order == "" {
		p.Order = "desc"
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 10 // default page size
	}
}
