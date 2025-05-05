package link

import (
	"errors"
	"net/http"
	"strconv"

	configs "UrlShortenerBackend/config"
	"UrlShortenerBackend/pkg/req"
	"UrlShortenerBackend/pkg/res"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type LinkHandlerDeps struct {
	LinkRepository *LinkRepository
	Config         *configs.Config
	Logger         *zerolog.Logger
}

type LinkHandler struct {
	LinkRepository *LinkRepository
	Logger         *zerolog.Logger
}

func NewLinkHandler(router *http.ServeMux, deps *LinkHandlerDeps) {
	handler := &LinkHandler{
		LinkRepository: deps.LinkRepository,
		Logger:         deps.Logger,
	}

	router.HandleFunc("GET /{hash}", handler.Redirect())

	router.HandleFunc("GET /api/v1/links", handler.GetLink())
	router.HandleFunc("GET /api/v1/links/all", handler.GetAllLinks())
	router.HandleFunc("POST /api/v1/links", handler.CreateLink())
	router.HandleFunc("DELETE /api/v1/links", handler.DeleteLink())
}

// Redirect godoc
// @Summary Redirect to original URL
// @Description Redirects to the original URL using the provided hash
// @Tags links
// @Produce html
// @Param hash path string true "Hash of the shortened link"
// @Success 302 {string} string "Redirect to the original URL"
// @Failure 400 {string} string "Hash parameter is missing"
// @Failure 404 {string} string "Link not found"
// @Router /{hash} [get]
func (handler *LinkHandler) Redirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue("hash")
		if hash == "" {
			handler.Logger.Error().Msg("Hash parameter is missing")
			http.Error(w, "Hash parameter is required", http.StatusBadRequest)
			return
		}

		link, err := handler.LinkRepository.GetLinkByHash(hash, "")
		if err != nil {
			handler.Logger.Error().Err(err).Str("hash", hash).Msg("Failed to find link by hash")
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}

		err = handler.LinkRepository.IncrementClicksCount(link)
		if err != nil {
			handler.Logger.Error().Err(err).Str("hash", hash).Msg("Failed to update click count")
		}

		handler.Logger.Info().
			Str("hash", hash).
			Str("url", link.Url).
			Int64("clicks", link.NumberOfClicks).
			Msg("Redirecting to URL")

		http.Redirect(w, r, link.Url, http.StatusFound)
	}
}

// GetLink godoc
// @Summary Get link details
// @Description Get details of a specific shortened link by hash
// @Tags links
// @Produce json
// @Param user_id query string true "User ID of the link owner"
// @Param hash query string true "Hash of the shortened link"
// @Success 200 {object} Link "Link details"
// @Failure 400 {string} string "Missing parameters"
// @Failure 403 {string} string "Link not found or user does not have access"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/links [get]
func (handler *LinkHandler) GetLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userId := r.URL.Query().Get("user_id")
		hash := r.URL.Query().Get("hash")

		if userId == "" {
			handler.Logger.Error().Msg("User ID is required")
			res.Json(w, "User ID is required", http.StatusBadRequest)
			return
		}

		if hash == "" {
			handler.Logger.Error().Msg("Hash is required")
			res.Json(w, "Hash is required", http.StatusBadRequest)
			return
		}

		link, err := handler.LinkRepository.GetLinkByHash(hash, userId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				handler.Logger.Error().
					Err(err).
					Str("hash", hash).
					Str("user_id", userId).
					Msg("Link not found or user does not have access")
				res.Json(w, "Link not found or user does not have access", http.StatusForbidden)
				return
			}

			handler.Logger.Error().Err(err).Str("hash", hash).Msg("Failed to find link")
			res.Json(w, "Failed to retrieve link", http.StatusInternalServerError)
			return
		}

		handler.Logger.Info().
			Str("hash", hash).
			Str("url", link.Url).
			Str("user_id", link.UserId).
			Msg("Link details retrieved successfully")

		res.Json(w, link, http.StatusOK)
	}
}

// GetAllLinks godoc
// @Summary Get all user links
// @Description Get a list of all links belonging to a user
// @Tags links
// @Produce json
// @Param user_id query string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} GetAllLinksResponse "List of links"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/links/all [get]
func (handler *LinkHandler) GetAllLinks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.URL.Query().Get("user_id")

		page := 1
		limit := 10

		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		if userId == "" {
			handler.Logger.Warn().Msg("User ID is missing, returning empty result")
			emptyResponse := GetAllLinksResponse{
				Links:      []Link{},
				TotalPages: 0,
				TotalLinks: 0,
				Page:       page,
				Limit:      limit,
			}
			res.Json(w, emptyResponse, http.StatusOK)
			return
		}

		exists, err := handler.LinkRepository.CheckUserExists(userId)
		if err != nil {
			handler.Logger.Error().Err(err).Str("user_id", userId).Msg("Failed to check user existence")
			res.Json(w, "Failed to check user existence", http.StatusInternalServerError)
			return
		}

		if !exists {
			handler.Logger.Info().Str("user_id", userId).Msg("User ID does not exist, returning empty result")
			emptyResponse := GetAllLinksResponse{
				Links:      []Link{},
				TotalPages: 0,
				TotalLinks: 0,
				Page:       page,
				Limit:      limit,
			}
			res.Json(w, emptyResponse, http.StatusOK)
			return
		}

		result, err := handler.LinkRepository.GetAllLinks(userId, page, limit)
		if err != nil {
			handler.Logger.Error().Err(err).Msg("Failed to get links")
			res.Json(w, err.Error(), http.StatusInternalServerError)
			return
		}

		handler.Logger.Info().
			Str("user_id", userId).
			Int("total_links", int(result.TotalLinks)).
			Int("page", result.Page).
			Int("total_pages", result.TotalPages).
			Msg("Successfully retrieved user links")

		response := GetAllLinksResponse{
			Links:      result.Links,
			TotalPages: result.TotalPages,
			TotalLinks: result.TotalLinks,
			Page:       result.Page,
			Limit:      result.Limit,
		}

		res.Json(w, response, http.StatusOK)
	}
}

// CreateLink godoc
// @Summary Create a new shortened link
// @Description Creates a new shortened link
// @Tags links
// @Accept json
// @Produce json
// @Param payload body LinkCreateRequest true "Data for creating a link"
// @Success 201 {object} Link "Created link"
// @Failure 400 {string} string "Error in request parameters"
// @Failure 404 {string} string "User ID not found"
// @Failure 409 {string} string "Hash already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/links [post]
func (handler *LinkHandler) CreateLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		payload, err := req.HandleBody[LinkCreateRequest](&w, r)
		if err != nil {
			handler.Logger.Error().Err(err).Msg("Failed to process create link request")
			return
		}

		if payload.UserId != "" {
			exists, err := handler.LinkRepository.CheckUserExists(payload.UserId)
			if err != nil {
				handler.Logger.Error().Err(err).Str("user_id", payload.UserId).Msg("Failed to check user existence")
				res.Json(w, "Failed to check user existence", http.StatusInternalServerError)
				return
			}

			if !exists {
				handler.Logger.Error().Str("user_id", payload.UserId).Msg("User ID does not exist")
				res.Json(w, "User ID not found", http.StatusNotFound)
				return
			}

			handler.Logger.Info().Str("user_id", payload.UserId).Msg("User ID exists, using it for the new link")
		}

		link := &Link{
			Url:            payload.Url,
			Hash:           payload.Hash,
			UserId:         payload.UserId,
			Lifetime:       DEFAULT_LIFETIME_DAYS,
			NumberOfClicks: DEFAULT_NUMBER_OF_CLICKS,
		}

		createdLink, err := handler.LinkRepository.Create(link)
		if err != nil {
			if err.Error() == "hash already exists" {
				handler.Logger.Warn().
					Str("url", payload.Url).
					Str("hash", payload.Hash).
					Msg("Attempted to create link with existing hash")
				res.Json(w, "Hash already exists", http.StatusConflict)
				return
			}

			handler.Logger.Error().
				Err(err).
				Str("url", payload.Url).
				Msg("Failed to create link")
			res.Json(w, "Failed to create link", http.StatusInternalServerError)
			return
		}

		handler.Logger.Info().
			Str("url", createdLink.Url).
			Str("hash", createdLink.Hash).
			Str("user_id", createdLink.UserId).
			Int64("lifetime", createdLink.Lifetime).
			Msg("Link created successfully")

		res.Json(w, createdLink, http.StatusCreated)
	}
}

// DeleteLink godoc
// @Summary Delete a shortened link
// @Description Deletes a shortened link by hash
// @Tags links
// @Accept json
// @Produce json
// @Param payload body LinkDeleteRequest true "Data for deleting a link"
// @Success 200 {string} string "Link deleted successfully"
// @Failure 400 {string} string "Error in request parameters"
// @Failure 403 {string} string "Link not found or user does not have permission"
// @Failure 404 {string} string "Link not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/links [delete]
func (handler *LinkHandler) DeleteLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := req.HandleBody[LinkDeleteRequest](&w, r)
		if err != nil {
			handler.Logger.Error().Err(err).Msg("Failed to process delete link request")
			return
		}

		if payload.Hash == "" {
			handler.Logger.Error().Msg("Hash is required")
			res.Json(w, "Hash is required", http.StatusBadRequest)
			return
		}

		if payload.UserId == "" {
			handler.Logger.Error().Msg("User ID is required")
			res.Json(w, "User ID is required", http.StatusBadRequest)
			return
		}

		err = handler.LinkRepository.DeleteLink(payload.Hash, payload.UserId)
		if err != nil {
			if err.Error() == "link not found or user does not have permission" {
				handler.Logger.Error().
					Err(err).
					Str("hash", payload.Hash).
					Str("user_id", payload.UserId).
					Msg("User does not have permission or link not found")
				res.Json(w, "Link not found or user does not have permission", http.StatusForbidden)
				return
			}

			if err.Error() == "link not found or already deleted" {
				handler.Logger.Warn().
					Str("hash", payload.Hash).
					Msg("Link not found or already deleted")
				res.Json(w, "Link not found", http.StatusNotFound)
				return
			}

			handler.Logger.Error().
				Err(err).
				Str("hash", payload.Hash).
				Msg("Failed to delete link")
			res.Json(w, "Failed to delete link", http.StatusInternalServerError)
			return
		}

		handler.Logger.Info().
			Str("hash", payload.Hash).
			Str("user_id", payload.UserId).
			Msg("Link deleted successfully")

		res.Json(w, "Link deleted successfully", http.StatusOK)
	}
}
