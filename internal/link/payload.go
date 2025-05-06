package link

type LinkCreateRequest struct {
	Url    string `json:"url" validate:"required,url" example:"https://example.com"`
	UserId string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Hash   string `json:"hash" example:"custom123"`
}

type LinkDeleteRequest struct {
	Hash   string `json:"hash" validate:"required" example:"abc123"`
	UserId string `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type GetLinkRequest struct {
	UserId string `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Hash   string `json:"hash" validate:"required" example:"abc123"`
}

type GetAllLinksResponse struct {
	Links      []Link `json:"links"`
	TotalPages int    `json:"total_pages" example:"5"`
	TotalLinks int64  `json:"total_links" example:"42"`
	Page       int    `json:"page" example:"1"`
	Limit      int    `json:"limit" example:"10"`
}

type GetAllLinksRequest struct {
	UserId string `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Page   int    `json:"page" example:"1"`
	Limit  int    `json:"limit" example:"10"`
}

type AddDaysRequest struct {
	UserId string `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}
