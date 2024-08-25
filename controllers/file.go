package controllers

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CanPacis/tstud-core/db"
	"github.com/gabriel-vasile/mimetype"
	"gorm.io/gorm"
)

type FileController struct {
	DB *gorm.DB
}

func NewFileController(db *gorm.DB) *FileController {
	return &FileController{DB: db}
}

func extractFiles(dir string, recursive bool, exclude []string) ([]db.File, error) {
	name := filepath.Base(dir)
	if slices.Contains(exclude, name) {
		return []db.File{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := []db.File{}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if slices.Contains(exclude, entry.Name()) {
			continue
		}

		if entry.IsDir() {
			if recursive {
				subFiles, err := extractFiles(path, recursive, exclude)
				if err != nil {
					return nil, err
				}

				files = append(files, subFiles...)
			}
		} else {
			file, err := extractFile(path)
			if err != nil {
				return nil, err
			}
			files = append(files, *file)
		}
	}

	return files, err
}

func extractFile(path string) (*db.File, error) {
	mtype, err := mimetype.DetectFile(path)
	var mime string
	if err == nil {
		mime = mtype.String()
	}

	return &db.File{
		FilePath: path,
		MimeType: mime,
	}, nil
}

func chunkFiles(files []db.File, capacity int) [][]db.File {
	chunks := [][]db.File{}
	for i := 0; i < len(files); i += capacity {
		end := i + capacity
		if end > len(files) {
			end = len(files)
		}
		chunk := files[i:end]
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (c *FileController) Index(path string, recursive bool, exclude []string) (*PaginatedResource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var files []db.File
	if info.IsDir() {
		var err error
		files, err = extractFiles(path, recursive, exclude)
		if err != nil {
			return nil, err
		}
	} else {
		file, err := extractFile(path)
		if err != nil {
			return nil, err
		}
		files = []db.File{*file}
	}

	chunks := chunkFiles(files, 20)

	for _, chunk := range chunks {
		// TODO: Known bug: if you index a directory, create a new file in that direactory and then reindex, this will throw because other resources violate unique constraints so the new file won't be indexed.
		tx := c.DB.Create(&chunk)
		if tx.Error != nil {
			if !(info.IsDir() && errors.Is(tx.Error, gorm.ErrDuplicatedKey)) {
				return nil, tx.Error
			}
		}
	}

	result := &PaginatedResource{}

	for _, file := range files {
		dto := *file.ToDTO()
		if dto.ID != 0 {
			result.Items = append(result.Items, *file.ToDTO())
		}
	}
	result.Page = 0
	result.TotalPages = 1

	return result, nil
}

func (c *FileController) Unindex(path string, recursive bool, exclude []string) (*PaginatedResource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var files []db.File
	if info.IsDir() {
		var err error
		files, err = extractFiles(path, recursive, exclude)
		if err != nil {
			return nil, err
		}
	} else {
		file, err := extractFile(path)
		if err != nil {
			return nil, err
		}
		files = []db.File{*file}
	}

	result := &PaginatedResource{}

	for _, file := range files {
		dto := *file.ToDTO()
		tx := c.DB.Delete(&file, "file_path = ?", file.FilePath)
		if tx.Error != nil {
			return nil, err
		}
		if dto.ID != 0 {
			result.Items = append(result.Items, *file.ToDTO())
		}
	}

	result.Page = 0
	result.TotalPages = 1

	return result, nil
}

func (c *FileController) Rename(oldPath, newPath string) (*db.FileDTO, error) {
	var file db.File
	tx := c.DB.First(&file, "file_path = ?", oldPath)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// TODO: maybe check the path? I don't know if something should exist or not exist there. Maybe just check if the path is valid.
	file.FilePath = newPath
	tx = c.DB.Save(&file)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return file.ToDTO(), nil
}

func (c *FileController) FindByID(id uint) (*db.FileDTO, error) {
	var file db.File
	tx := c.DB.Preload("Tags").First(&file, "id = ?", id)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return file.ToDTO(), tx.Error
}

func (c *FileController) FindByPath(path string) (*db.FileDTO, error) {
	var file db.File
	tx := c.DB.Preload("Tags").First(&file, "file_path = ?", path)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return file.ToDTO(), tx.Error
}

func (c *FileController) Tag(fileId uint, tagId uint) (*db.FileDTO, *db.TagDTO, error) {
	var file db.File
	var tag db.Tag

	tx := c.DB.First(&file, "id = ?", fileId)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	c.DB.First(&tag, "id = ?", tagId)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	return file.ToDTO(), tag.ToDTO(), c.DB.Model(&file).Association("Tags").Append(&tag)
}

func (c *FileController) Untag(fileId uint, tagId uint) (*db.FileDTO, *db.TagDTO, error) {
	var file db.File
	var tag db.Tag

	tx := c.DB.First(&file, "id = ?", fileId)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	c.DB.First(&tag, "id = ?", tagId)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	return file.ToDTO(), tag.ToDTO(), c.DB.Model(&file).Association("Tags").Delete(&tag)
}

type FileMetaData struct {
	Author      *string
	Description *string
}

func (c *FileController) SetMeta(fileId uint, meta FileMetaData) (*db.FileDTO, error) {
	var file db.File

	tx := c.DB.First(&file, "id = ?", fileId)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// TODO: update this to be more dynamic
	if meta.Author != nil {
		file.Author = *meta.Author
	}

	if meta.Description != nil {
		file.Description = *meta.Description
	}

	tx = c.DB.Save(&file)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return file.ToDTO(), nil
}

func (c *FileController) List(options ListOptions) (*PaginatedResource, error) {
	var files []db.File

	tx := c.DB.Order("file_path desc").Limit(options.PerPage).Offset(options.Page * options.PerPage).Find(&files)
	if tx.Error != nil {
		return nil, tx.Error
	}

	var count int64
	tx = c.DB.Model(&db.File{}).Count(&count)
	if tx.Error != nil {
		return nil, tx.Error
	}

	result := &PaginatedResource{
		Items:      []any{},
		Page:       options.Page,
		TotalPages: int(math.Ceil(float64(count) / float64(options.PerPage))),
	}

	for _, file := range files {
		result.Items = append(result.Items, *file.ToDTO())
	}

	return result, nil
}

func (c *FileController) extractTags(tagNames []string) ([]uint, error) {
	var tags []db.Tag

	tx := c.DB.Find(&tags, "tag_name in ?", tagNames)
	if tx.Error != nil {
		return nil, tx.Error
	}

	tagIds := []uint{}
	for _, tag := range tags {
		tagIds = append(tagIds, tag.ID)
	}

	return c.extractTagIDs(tagIds, 0)
}

func (c *FileController) extractTagIDs(ids []uint, depth int) ([]uint, error) {
	if depth > 6 || len(ids) == 0 {
		return []uint{}, nil
	}

	var tags []db.Tag

	tagIds := []uint{}
	if depth == 0 {
		tagIds = append(tagIds, ids...)
	}

	tx := c.DB.Find(&tags, "parent_id in ?", ids)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, tag := range tags {
		tagIds = append(tagIds, tag.ID)
	}

	subTags, err := c.extractTagIDs(tagIds, depth+1)
	if err != nil {
		return nil, err
	}
	tagIds = append(tagIds, subTags...)

	return tagIds, nil
}

func (c *FileController) tagSearch(tagIds []uint, limit, offset int) ([]db.File, error) {
	if len(tagIds) == 0 {
		return []db.File{}, nil
	}

	var fileTags []struct {
		FileID uint
		TagID  uint
	}
	tx := c.DB.Table("file_tags").Find(&fileTags, "tag_id IN ?", tagIds)
	if tx.Error != nil {
		return nil, tx.Error
	}

	var fileIds []uint
	for _, fileTag := range fileTags {
		fileIds = append(fileIds, fileTag.FileID)
	}

	var files []db.File
	tx = c.DB.Order("file_path desc").Limit(limit).Offset(offset).Find(&files, "id IN ?", fileIds)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return files, nil
}

func (c *FileController) aliasSearch(words []string, limit, offset int) ([]db.File, error) {
	var aliases []db.Alias

	tx := c.DB.Find(&aliases, "name IN ?", words)
	if tx.Error != nil {
		return nil, tx.Error
	}

	var tagIds []uint

	for _, alias := range aliases {
		tagIds = append(tagIds, alias.TagID)
	}

	return c.tagSearch(tagIds, limit, offset)
}

// TODO: Fix this to do a proper keyword sql like search
func (c *FileController) keywordSearch(keywords []string, limit, offset int) ([]db.File, error) {
	files := []db.File{}
	tx := c.DB.Limit(limit).Offset(offset).Find(&files, "file_name IN ?", keywords)

	return files, tx.Error
}

func (c *FileController) Search(options SearchOptions) (*PaginatedResource, error) {
	result := &PaginatedResource{
		Items:      []any{},
		Page:       options.Page,
		TotalPages: 1,
	}

	files := []db.File{}

	// TODO: Make this concurrent
	tagIds, err := c.extractTags(options.Tags)
	if err != nil {
		return nil, err
	}
	search, err := c.tagSearch(tagIds, options.PerPage, options.PerPage*options.Page)
	if err == nil {
		files = append(files, search...)
	}
	search, err = c.aliasSearch(strings.Split(options.Term, " "), options.PerPage, options.PerPage*options.Page)
	if err == nil {
		files = append(files, search...)
	}
	search, err = c.keywordSearch(strings.Split(options.Term, " "), options.PerPage, options.PerPage*options.Page)
	if err == nil {
		files = append(files, search...)
	}

	result.TotalPages = len(files)/options.PerPage + 1

	for _, file := range files {
		result.Items = append(result.Items, *file.ToDTO())
	}

	return result, nil
}
