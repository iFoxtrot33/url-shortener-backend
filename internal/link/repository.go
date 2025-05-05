package link

import (
	"UrlShortenerBackend/pkg/db"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaginationResult struct {
	Links      []Link
	TotalLinks int64
	TotalPages int
	Page       int
	Limit      int
}

type LinkRepository struct {
	Database *db.Db
}

func NewLinkRepository(database *db.Db) *LinkRepository {
	return &LinkRepository{
		Database: database,
	}
}

func (repo *LinkRepository) GetAllLinks(userId string, page, limit int) (*PaginationResult, error) {
	if userId == "" {
		return &PaginationResult{
			Links:      []Link{},
			TotalLinks: 0,
			TotalPages: 0,
			Page:       page,
			Limit:      limit,
		}, nil
	}

	if page <= 0 {
		page = DEFAULT_PAGE
	}

	if limit <= 0 {
		limit = DEFAULT_LIMIT
	}

	offset := (page - 1) * limit

	var links []Link
	result := repo.Database.DB.Table("links").Where("deleted_at IS NULL AND user_id = ?", userId).Order("created_at DESC").Offset(offset).Limit(limit).Scan(&links)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(links) == 0 {
		return &PaginationResult{
			Links:      []Link{},
			TotalLinks: 0,
			TotalPages: 0,
			Page:       page,
			Limit:      limit,
		}, nil
	}

	var totalLinks int64
	countResult := repo.Database.DB.Table("links").Where("deleted_at IS NULL AND user_id = ?", userId).Count(&totalLinks)
	if countResult.Error != nil {
		return nil, countResult.Error
	}

	totalPages := (int(totalLinks) + limit - 1) / limit

	return &PaginationResult{
		Links:      links,
		TotalLinks: totalLinks,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (repo *LinkRepository) GetLinkByHash(hash string, userId string) (*Link, error) {
	if userId != "" {
		var link Link
		result := repo.Database.DB.Where("hash = ? AND user_id = ? AND deleted_at IS NULL", hash, userId).First(&link)
		if result.Error != nil {
			return nil, result.Error
		}
		return &link, nil
	}

	var link Link
	result := repo.Database.DB.Where("hash = ? AND deleted_at IS NULL", hash).First(&link)
	if result.Error != nil {
		return nil, result.Error
	}

	return &link, nil
}

func (repo *LinkRepository) Create(link *Link) (*Link, error) {
	if link.Hash == "" {
		link.Hash = RandStringRunes(10)

		for {
			var existingLink Link
			result := repo.Database.DB.Where("hash = ? AND deleted_at IS NULL", link.Hash).First(&existingLink)

			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				break
			}

			if result.Error != nil {
				return nil, fmt.Errorf("error checking hash uniqueness: %w", result.Error)
			}

			link.Hash = RandStringRunes(10)
		}
	} else {
		var existingLink Link
		result := repo.Database.DB.Where("hash = ? AND deleted_at IS NULL", link.Hash).First(&existingLink)

		if result.Error == nil {
			return nil, errors.New("hash already exists")
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("error checking hash existence: %w", result.Error)
		}
	}

	if link.UserId == "" {
		link.UserId = uuid.New().String()

		for {
			var existingLink Link
			result := repo.Database.DB.Where("user_id = ?", link.UserId).First(&existingLink)

			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				break
			}

			if result.Error != nil {
				return nil, fmt.Errorf("error checking user ID uniqueness: %w", result.Error)
			}

			link.UserId = uuid.New().String()
		}
	}

	var deletedLink Link
	deletedResult := repo.Database.DB.Unscoped().Where("hash = ? AND deleted_at IS NOT NULL", link.Hash).First(&deletedLink)

	if deletedResult.Error == nil {
		deleteErr := repo.Database.DB.Unscoped().Delete(&deletedLink).Error
		if deleteErr != nil {
			return nil, fmt.Errorf("error removing previously deleted link with same hash: %w", deleteErr)
		}
	}

	result := repo.Database.DB.Create(link)
	if result.Error != nil {
		return nil, fmt.Errorf("error creating link: %w", result.Error)
	}

	return link, nil
}

func (repo *LinkRepository) DecrementLifetimeForAllLinks() error {

	result := repo.Database.DB.Exec("UPDATE links SET lifetime = lifetime - 1 WHERE lifetime > 0 AND deleted_at IS NULL")
	return result.Error
}

func (repo *LinkRepository) DeleteExpiredLinks() error {

	var expiredLinks []Link
	result := repo.Database.DB.Where("lifetime = 0 AND deleted_at IS NULL").Find(&expiredLinks)
	if result.Error != nil {
		return result.Error
	}

	count := len(expiredLinks)
	if count == 0 {
		return nil
	}

	result = repo.Database.DB.Delete(&expiredLinks)

	return result.Error
}

func (repo *LinkRepository) IncrementClicksCount(link *Link) error {
	link.NumberOfClicks++
	result := repo.Database.DB.Model(link).Update("number_of_clicks", link.NumberOfClicks)
	return result.Error
}

func (repo *LinkRepository) DeleteLink(hash string, userId string) error {
	matches, err := repo.CheckUserMatchesLink(hash, userId)
	if err != nil {
		return err
	}

	if !matches {
		return errors.New("link not found or user does not have permission")
	}

	return repo.Database.DB.Transaction(func(tx *gorm.DB) error {
		var link Link
		if err := tx.Where("hash = ? AND user_id = ? AND deleted_at IS NULL", hash, userId).First(&link).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("link not found or already deleted")
			}
			return err
		}

		if err := tx.Delete(&link).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *LinkRepository) CheckUserExists(userId string) (bool, error) {
	if userId == "" {
		return false, nil
	}

	var count int64
	result := repo.Database.DB.Table("links").Where("user_id = ? AND deleted_at IS NULL", userId).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

func (repo *LinkRepository) CheckUserMatchesLink(hash, userId string) (bool, error) {
	if hash == "" || userId == "" {
		return false, errors.New("hash and userId are required")
	}

	var link Link
	result := repo.Database.DB.Where("hash = ? AND user_id = ? AND deleted_at IS NULL", hash, userId).First(&link)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}
