package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ZygmaCore/kids_planet/services/api/internal/clients"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type GameService struct {
	gameRepo    *repos.GameRepo
	minio       *clients.MinIO
	minioBucket string
	zipMaxBytes int64
}

func NewGameService(gameRepo *repos.GameRepo, minio *clients.MinIO, minioBucket string, zipMaxBytes int64) *GameService {
	return &GameService{
		gameRepo:    gameRepo,
		minio:       minio,
		minioBucket: strings.TrimSpace(minioBucket),
		zipMaxBytes: zipMaxBytes,
	}
}

type ListPublicGamesInput struct {
	AgeCategoryID       *int64
	EducationCategoryID *int64
	Sort                string
	Page                int
	Limit               int
}

type GameListDTO struct {
	Items []GameListItemDTO `json:"items"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Total int               `json:"total"`
}

type GameListItemDTO struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Thumbnail     *string `json:"thumbnail"`
	GameURL       *string `json:"game_url"`
	AgeCategoryID int64   `json:"age_category_id"`
	Free          bool    `json:"free"`
	CreatedAt     string  `json:"created_at"`
}

type GameDetailDTO struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Thumbnail     *string `json:"thumbnail"`
	GameURL       *string `json:"game_url"`
	AgeCategoryID int64   `json:"age_category_id"`
	Free          bool    `json:"free"`
	CreatedAt     string  `json:"created_at"`
}

func (s *GameService) ListPublicGames(ctx context.Context, in ListPublicGamesInput) (*GameListDTO, error) {
	page := in.Page
	limit := in.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 24
	}

	if page < 1 {
		return nil, utils.ErrBadRequest("page must be >= 1")
	}
	if limit < 1 || limit > 100 {
		return nil, utils.ErrBadRequest("limit must be between 1 and 100")
	}

	sort := strings.TrimSpace(strings.ToLower(in.Sort))
	if sort == "" {
		sort = "newest"
	}
	var sortEnum repos.GameListSort
	switch sort {
	case "newest":
		sortEnum = repos.GameSortNewest
	case "popular":
		sortEnum = repos.GameSortPopular
	default:
		return nil, utils.ErrBadRequest("sort must be one of: newest, popular")
	}

	var age sql.NullInt64
	if in.AgeCategoryID != nil {
		if *in.AgeCategoryID < 1 {
			return nil, utils.ErrBadRequest("age_category_id must be >= 1")
		}
		age = sql.NullInt64{Int64: *in.AgeCategoryID, Valid: true}
	}

	var edu sql.NullInt64
	if in.EducationCategoryID != nil {
		if *in.EducationCategoryID < 1 {
			return nil, utils.ErrBadRequest("education_category_id must be >= 1")
		}
		edu = sql.NullInt64{Int64: *in.EducationCategoryID, Valid: true}
	}

	filter := repos.GameListFilter{
		AgeCategoryID:       age,
		EducationCategoryID: edu,
		Sort:                sortEnum,
		Page:                page,
		Limit:               limit,
	}

	items, err := s.gameRepo.ListPublic(ctx, filter)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	total, err := s.gameRepo.CountPublic(ctx, filter)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	out := make([]GameListItemDTO, 0, len(items))
	for _, it := range items {
		var thumb *string
		if it.Thumbnail.Valid {
			v := it.Thumbnail.String
			thumb = &v
		}

		var url *string
		if it.GameURL.Valid {
			v := it.GameURL.String
			url = &v
		}

		out = append(out, GameListItemDTO{
			ID:            it.ID,
			Title:         it.Title,
			Slug:          it.Slug,
			Thumbnail:     thumb,
			GameURL:       url,
			AgeCategoryID: it.AgeCategoryID,
			Free:          it.Free,
			CreatedAt:     it.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return &GameListDTO{
		Items: out,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *GameService) GetPublicGameByID(ctx context.Context, id int64) (*GameDetailDTO, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}

	it, err := s.gameRepo.GetByIDPublic(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	var thumb *string
	if it.Thumbnail.Valid {
		v := it.Thumbnail.String
		thumb = &v
	}

	var url *string
	if it.GameURL.Valid {
		v := it.GameURL.String
		url = &v
	}

	return &GameDetailDTO{
		ID:            it.ID,
		Title:         it.Title,
		Slug:          it.Slug,
		Thumbnail:     thumb,
		GameURL:       url,
		AgeCategoryID: it.AgeCategoryID,
		Free:          it.Free,
		CreatedAt:     it.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type AdminListGamesInput struct {
	Status string
	Q      string
	Page   int
	Limit  int
}

type AdminGameListDTO struct {
	Items []models.AdminGameDTO `json:"items"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
	Total int                   `json:"total"`
}

func (s *GameService) ListAdminGames(ctx context.Context, in AdminListGamesInput) (*AdminGameListDTO, error) {
	page := in.Page
	limit := in.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 24
	}
	if page < 1 {
		return nil, utils.ErrBadRequest("page must be >= 1")
	}
	if limit < 1 || limit > 100 {
		return nil, utils.ErrBadRequest("limit must be between 1 and 100")
	}

	status := strings.TrimSpace(strings.ToLower(in.Status))
	var statusNS sql.NullString
	if status != "" {
		if status != "draft" && status != "active" && status != "archived" {
			return nil, utils.ErrBadRequest("status must be one of: draft, active, archived")
		}
		statusNS = sql.NullString{String: status, Valid: true}
	}

	q := strings.TrimSpace(in.Q)
	var qNS sql.NullString
	if q != "" {
		qNS = sql.NullString{String: q, Valid: true}
	}

	filter := repos.AdminGameFilter{
		Status: statusNS,
		Q:      qNS,
		Page:   page,
		Limit:  limit,
	}

	items, err := s.gameRepo.ListAdmin(ctx, filter)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	total, err := s.gameRepo.CountAdmin(ctx, filter)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	out := make([]models.AdminGameDTO, 0, len(items))
	for _, g := range items {
		out = append(out, toAdminGameDTO(g))
	}

	return &AdminGameListDTO{
		Items: out,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *GameService) CreateAdminGame(ctx context.Context, createdBy int64, req models.CreateGameRequest) (*models.AdminGameDTO, error) {
	title := strings.TrimSpace(req.Title)
	slug := strings.TrimSpace(req.Slug)

	if title == "" {
		return nil, utils.ErrBadRequest("title is required")
	}
	if len(title) > 150 {
		return nil, utils.ErrBadRequest("title must be <= 150 chars")
	}
	if slug == "" {
		return nil, utils.ErrBadRequest("slug is required")
	}
	if len(slug) > 150 {
		return nil, utils.ErrBadRequest("slug must be <= 150 chars")
	}
	if !slugRe.MatchString(slug) {
		return nil, utils.ErrBadRequest("slug must be lowercase and dash-separated (e.g. color-match)")
	}
	if req.AgeCategoryID < 1 {
		return nil, utils.ErrBadRequest("age_category_id must be >= 1")
	}
	if createdBy <= 0 {
		return nil, utils.ErrInternal()
	}

	ok, err := s.gameRepo.AgeCategoryExists(ctx, req.AgeCategoryID)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	if !ok {
		return nil, utils.ErrBadRequest("age_category_id not found")
	}

	exists, err := s.gameRepo.SlugExists(ctx, slug, nil)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	if exists {
		return nil, utils.ErrBadRequest("slug already exists")
	}

	free := true
	if req.Free != nil {
		free = *req.Free
	}

	var thumb sql.NullString
	if strings.TrimSpace(req.Thumbnail) != "" {
		thumb = sql.NullString{String: strings.TrimSpace(req.Thumbnail), Valid: true}
	}
	var gameURL sql.NullString
	if strings.TrimSpace(req.GameURL) != "" {
		gameURL = sql.NullString{String: strings.TrimSpace(req.GameURL), Valid: true}
	}

	g, err := s.gameRepo.CreateAdminGame(ctx, repos.CreateAdminGameInput{
		Title:         title,
		Slug:          slug,
		Thumbnail:     thumb,
		GameURL:       gameURL,
		AgeCategoryID: req.AgeCategoryID,
		Free:          free,
		CreatedBy:     createdBy,
	})
	if err != nil {
		return nil, utils.ErrInternal()
	}

	dto := toAdminGameDTO(*g)
	return &dto, nil
}

func (s *GameService) UpdateAdminGame(ctx context.Context, id int64, req models.UpdateGameRequest) (*models.AdminGameDTO, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}

	_, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, utils.ErrBadRequest("title cannot be empty")
		}
		if len(t) > 150 {
			return nil, utils.ErrBadRequest("title must be <= 150 chars")
		}
		req.Title = &t
	}

	if req.Slug != nil {
		slug := strings.TrimSpace(*req.Slug)
		if slug == "" {
			return nil, utils.ErrBadRequest("slug cannot be empty")
		}
		if len(slug) > 150 {
			return nil, utils.ErrBadRequest("slug must be <= 150 chars")
		}
		if !slugRe.MatchString(slug) {
			return nil, utils.ErrBadRequest("slug must be lowercase and dash-separated (e.g. color-match)")
		}

		exists, err := s.gameRepo.SlugExists(ctx, slug, &id)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		if exists {
			return nil, utils.ErrBadRequest("slug already exists")
		}
		req.Slug = &slug
	}

	if req.AgeCategoryID != nil {
		if *req.AgeCategoryID < 1 {
			return nil, utils.ErrBadRequest("age_category_id must be >= 1")
		}
		ok, err := s.gameRepo.AgeCategoryExists(ctx, *req.AgeCategoryID)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		if !ok {
			return nil, utils.ErrBadRequest("age_category_id not found")
		}
	}

	g, err := s.gameRepo.UpdateAdminGame(ctx, id, repos.UpdateAdminGameInput{
		Title:         req.Title,
		Slug:          req.Slug,
		Thumbnail:     req.Thumbnail,
		GameURL:       req.GameURL,
		AgeCategoryID: req.AgeCategoryID,
		Free:          req.Free,
	})
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	dto := toAdminGameDTO(*g)
	return &dto, nil
}

func (s *GameService) PublishAdminGame(ctx context.Context, id int64) (*models.AdminGameDTO, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}

	g0, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}
	if g0.Status == "archived" {
		return nil, utils.ErrBadRequest("archived game cannot be published")
	}

	g, err := s.gameRepo.SetStatus(ctx, id, "active")
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	dto := toAdminGameDTO(*g)
	return &dto, nil
}

func (s *GameService) UnpublishAdminGame(ctx context.Context, id int64) (*models.AdminGameDTO, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}

	g0, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}
	if g0.Status == "archived" {
		return nil, utils.ErrBadRequest("archived game cannot be unpublished")
	}

	g, err := s.gameRepo.SetStatus(ctx, id, "draft")
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	dto := toAdminGameDTO(*g)
	return &dto, nil
}

type UploadZipDTO struct {
	ObjectKey string `json:"object_key"`
	ETag      string `json:"etag"`
	Size      int64  `json:"size"`
	GameURL   string `json:"game_url"`
}

func (s *GameService) UploadAdminGameZip(ctx context.Context, gameID int64, filename string, file io.ReadSeeker, size int64, contentType string) (*UploadZipDTO, error) {
	if gameID < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}
	if s.minio == nil {
		return nil, utils.ErrInternal()
	}
	if strings.TrimSpace(s.minioBucket) == "" {
		return nil, utils.ErrInternal()
	}
	if size <= 0 {
		return nil, utils.ErrBadRequest("file is required")
	}
	if s.zipMaxBytes > 0 && size > s.zipMaxBytes {
		return nil, utils.ErrZipTooLarge(s.zipMaxBytes)
	}

	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	if ext != ".zip" {
		return nil, utils.ErrInvalidZip("file must be a .zip")
	}

	g, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}
	if g.Status == "archived" {
		return nil, utils.ErrBadRequest("archived game cannot be uploaded")
	}

	head := make([]byte, 4)
	n, err := io.ReadFull(file, head)
	if err != nil || n < 2 {
		return nil, utils.ErrInvalidZip("invalid zip file")
	}
	_, _ = file.Seek(0, io.SeekStart)
	if head[0] != 'P' || head[1] != 'K' {
		return nil, utils.ErrInvalidZip("invalid zip file")
	}

	ct := strings.TrimSpace(contentType)
	if ct == "" {
		ct = "application/zip"
	}

	workDir, err := os.MkdirTemp("", "kids-planet-zip-")
	if err != nil {
		return nil, utils.ErrInternal()
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	zipPath := filepath.Join(workDir, "upload.zip")
	tmpFile, err := os.Create(zipPath)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		_ = tmpFile.Close()
		return nil, utils.ErrInternal()
	}
	if _, err := io.CopyN(tmpFile, file, size); err != nil {
		_ = tmpFile.Close()
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, utils.ErrInvalidZip("invalid zip file")
		}
		return nil, utils.ErrInternal()
	}
	if err := tmpFile.Close(); err != nil {
		return nil, utils.ErrInternal()
	}

	zipFile, err := os.Open(zipPath)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	defer func() { _ = zipFile.Close() }()

	info, err := zipFile.Stat()
	if err != nil {
		return nil, utils.ErrInternal()
	}

	extractDir := filepath.Join(workDir, "extracted")
	extracted, err := utils.SafeUnzip(zipFile, info.Size(), extractDir)
	if err != nil {
		var zipErr utils.ZipInputError
		if errors.As(err, &zipErr) {
			return nil, utils.ErrInvalidZip(zipErr.Error())
		}
		return nil, utils.ErrInternal()
	}
	if !hasRootIndex(extracted) {
		return nil, utils.ErrMissingIndexHTML()
	}

	if err := s.uploadExtractedGameFiles(ctx, gameID, extractDir, extracted); err != nil {
		return nil, utils.ErrInternal()
	}

	now := time.Now().UTC()
	ts := now.Format("20060102_150405")
	rnd, err := randHex(8)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	uploadPrefix := fmt.Sprintf("%d/upload", gameID)
	objectKey := path.Join(uploadPrefix, fmt.Sprintf("%s_%s.zip", ts, rnd))

	if _, err := zipFile.Seek(0, io.SeekStart); err != nil {
		return nil, utils.ErrInternal()
	}
	etag, err := s.minio.PutObject(ctx, s.minioBucket, objectKey, zipFile, info.Size(), ct)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	playableURL := fmt.Sprintf("/games/%d/current/index.html", gameID)
	if _, err := s.gameRepo.SetGameURL(ctx, gameID, playableURL); err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	return &UploadZipDTO{
		ObjectKey: objectKey,
		ETag:      etag,
		Size:      info.Size(),
		GameURL:   playableURL,
	}, nil
}

func hasRootIndex(paths []string) bool {
	for _, p := range paths {
		if filepath.ToSlash(p) == "index.html" {
			return true
		}
	}
	return false
}

func (s *GameService) uploadExtractedGameFiles(ctx context.Context, gameID int64, root string, files []string) error {
	currentPrefix := fmt.Sprintf("%d/current", gameID)

	for _, rel := range files {
		rel = filepath.ToSlash(rel)
		if rel == "" {
			continue
		}

		fullPath := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Stat(fullPath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			continue
		}

		f, err := os.Open(fullPath)
		if err != nil {
			return err
		}

		contentType := mime.TypeByExtension(filepath.Ext(rel))
		if contentType == "" {
			head := make([]byte, 512)
			n, _ := f.Read(head)
			contentType = http.DetectContentType(head[:n])
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				_ = f.Close()
				return err
			}
		}

		objectKey := path.Join(currentPrefix, rel)
		if _, err := s.minio.PutObject(ctx, s.minioBucket, objectKey, f, info.Size(), contentType); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()
	}

	return nil
}

func randHex(nBytes int) (string, error) {
	if nBytes <= 0 {
		return "", errors.New("nBytes must be > 0")
	}
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toAdminGameDTO(g repos.Game) models.AdminGameDTO {
	var thumb string
	if g.Thumbnail.Valid {
		thumb = g.Thumbnail.String
	}
	var url string
	if g.GameURL.Valid {
		url = g.GameURL.String
	}

	return models.AdminGameDTO{
		ID:            g.ID,
		Title:         g.Title,
		Slug:          g.Slug,
		Status:        models.GameStatus(g.Status),
		Thumbnail:     thumb,
		GameURL:       url,
		AgeCategoryID: g.AgeCategoryID,
		Free:          g.Free,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}
