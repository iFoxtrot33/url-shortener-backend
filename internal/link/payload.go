package link

// LinkCreateRequest request for creating a new shortened link
type LinkCreateRequest struct {
	Url    string `json:"url" validate:"required,url" example:"https://example.com"` // Original URL to shorten
	UserId string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`    // User ID (optional)
	Hash   string `json:"hash" example:"custom123"`                                  // Custom hash (optional)
}

// LinkDeleteRequest request for deleting a shortened link
type LinkDeleteRequest struct {
	Hash   string `json:"hash" validate:"required" example:"abc123"`                                  // Hash of the link to delete
	UserId string `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"` // ID of the user who owns the link
}

// GetLinkRequest request for getting link information
type GetLinkRequest struct {
	UserId string `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"` // ID of the user who owns the link
	Hash   string `json:"hash" validate:"required" example:"abc123"`                                  // Hash of the link to get information
}

// GetAllLinksResponse response with a list of links
type GetAllLinksResponse struct {
	Links      []Link `json:"links"`                    // List of links
	TotalPages int    `json:"total_pages" example:"5"`  // Total number of pages
	TotalLinks int64  `json:"total_links" example:"42"` // Total number of links
	Page       int    `json:"page" example:"1"`         // Current page
	Limit      int    `json:"limit" example:"10"`       // Links per page limit
}

// GetAllLinksRequest request for getting a list of user's links
type GetAllLinksRequest struct {
	UserId string `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"` // User ID
	Page   int    `json:"page" example:"1"`                                                           // Page number
	Limit  int    `json:"limit" example:"10"`                                                         // Links per page limit
}
