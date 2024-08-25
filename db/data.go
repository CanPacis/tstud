package db

import (
	"path/filepath"

	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	FilePath    string `gorm:"unique;index"`
	MimeType    string
	Description string
	Author      string
	State       int   `gorm:"check: state IN (0,1);default: 0"`
	Tags        []Tag `gorm:"many2many:file_tags;"`
}

func (f File) Name() string {
	return filepath.Base(f.FilePath)
}

func (f File) ToDTO() *FileDTO {
	tags := []TagDTO{}

	for _, tag := range f.Tags {
		tags = append(tags, *tag.ToDTO())
	}

	return &FileDTO{
		ID:          f.ID,
		FilePath:    f.FilePath,
		Name:        f.Name(),
		MimeType:    f.MimeType,
		Description: f.Description,
		Author:      f.Author,
		Tags:        tags,
	}
}

type Tag struct {
	gorm.Model
	TagName  string
	ParentID *int
	Parent   *Tag `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Aliases  []Alias
}

func (t Tag) ToDTO() *TagDTO {
	var parent *TagDTO

	if t.Parent != nil {
		parent = t.Parent.ToDTO()
	}

	return &TagDTO{
		ID:      t.ID,
		Name:    t.TagName,
		Parent:  parent,
		Aliases: t.Aliases,
	}
}

type Alias struct {
	gorm.Model
	Name  string
	TagID uint
}

type FileDTO struct {
	ID          uint     `json:"id"`
	FilePath    string   `json:"file_path"`
	Name        string   `json:"name"`
	MimeType    string   `json:"mime_type"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Tags        []TagDTO `json:"tags"`
}

type TagDTO struct {
	ID      uint    `json:"id"`
	Name    string  `json:"name"`
	Parent  *TagDTO `json:"parent"`
	Aliases []Alias `json:"aliases"`
}

// func (dto TagDTO) String() string {
// 	if len(dto.Aliases) > 0 {
// 		aliases := []string{}
// 		for _, alias := range dto.Aliases {
// 			aliases = append(aliases, alias.Name)
// 		}
// 		return fmt.Sprintf("%d %s (%s)", dto.ID, dto.Name, strings.Join(aliases, ", "))
// 	}

// 	return fmt.Sprintf("%d %s", dto.ID, dto.Name)
// }
