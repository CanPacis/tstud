package controllers

import (
	"fmt"

	"github.com/CanPacis/tstud-core/db"
	"gorm.io/gorm"
)

type TagController struct {
	DB *gorm.DB
}

func NewTagController(db *gorm.DB) *TagController {
	return &TagController{DB: db}
}

func (c *TagController) Create(name string, parent *int) (*db.TagDTO, error) {
	tag := db.Tag{
		TagName:  name,
		ParentID: parent,
	}
	tx := c.DB.Create(&tag)
	return tag.ToDTO(), tx.Error
}

func (c *TagController) Delete(id uint) (*db.TagDTO, error) {
	var tag db.Tag
	tx := c.DB.First(&tag, "id = ?", id)
	if tx.Error != nil {
		return nil, tx.Error
	}

	tx = c.DB.Delete(&tag)
	return tag.ToDTO(), tx.Error
}

func (c *TagController) DeleteByName(name string) error {
	var tag db.Tag

	tx := c.DB.First(&tag, "tag_name = ?", name)
	if tx.Error != nil {
		return tx.Error
	}

	tx = c.DB.Delete(&tag)
	return tx.Error
}

func (c *TagController) Alias(id uint, alias string) error {
	var tag db.Tag

	tx := c.DB.First(&tag, "id = ?", id)
	if tx.Error != nil {
		return tx.Error
	}

	return c.DB.Model(&tag).Association("Aliases").Append(&db.Alias{Name: alias})
}

func (c *TagController) Unlias(id uint, aliasName string) error {
	var tag db.Tag
	var alias db.Alias

	tx := c.DB.Preload("Aliases").First(&tag, "id = ?", id)
	if tx.Error != nil {
		return tx.Error
	}

	for _, a := range tag.Aliases {
		if a.Name == aliasName {
			alias = a
		}
	}

	if alias.ID == 0 {
		return fmt.Errorf("tag %s does not have any alias named %s", tag.TagName, aliasName)
	}

	tx = c.DB.Delete(&alias)
	return tx.Error
}

func (c *TagController) List(parent *int) (*PaginatedResource, error) {
	var tags []db.Tag

	var tx *gorm.DB

	if parent == nil {
		tx = c.DB.
			Order("tag_name desc").
			Preload("Parent").
			Preload("Aliases").
			Find(&tags, "parent_id IS NULL")
	} else {
		if *parent < 0 {
			tx = c.DB.
				Order("tag_name desc").
				Preload("Parent").
				Preload("Aliases").
				Find(&tags)
		} else {
			tx = c.DB.
				Order("tag_name desc").
				Preload("Parent").
				Preload("Aliases").
				Find(&tags, "parent_id = ?", *parent)
		}
	}

	if tx.Error != nil {
		return nil, tx.Error
	}

	var count int64
	tx = c.DB.Model(&db.Tag{}).Count(&count)
	if tx.Error != nil {
		return nil, tx.Error
	}

	result := &PaginatedResource{
		Items:      []any{},
		Page:       0,
		TotalPages: 1,
	}

	for _, tag := range tags {
		result.Items = append(result.Items, *tag.ToDTO())
	}

	return result, nil
}

func (c *TagController) Search(term string, options ListOptions) (*PaginatedResource, error) {
	var tags []db.Tag
	query := c.DB.Preload("Parent").Preload("Aliases").Limit(options.PerPage).Offset(options.PerPage * options.Page)
	tx := query.Find(&tags, "tag_name LIKE ?", fmt.Sprintf("%%%s%%", term))
	if tx.Error != nil {
		return nil, tx.Error
	}

	items := []any{}
	for _, tag := range tags {
		items = append(items, *tag.ToDTO())
	}

	var aliases []db.Alias
	tx = c.DB.Find(&aliases, "name LIKE ?", fmt.Sprintf("%%%s%%", term))
	if tx.Error != nil {
		return nil, tx.Error
	}

	aliasedTagIds := []uint{}
	for _, alias := range aliases {
		aliasedTagIds = append(aliasedTagIds, alias.TagID)
	}

	query = c.DB.Preload("Parent").Preload("Aliases").Limit(options.PerPage).Offset(options.PerPage * options.Page)
	tx = query.Find(&tags, "id IN ?", aliasedTagIds)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, tag := range tags {
		items = append(items, *tag.ToDTO())
	}

	result := &PaginatedResource{
		Items:      items,
		Page:       options.Page,
		TotalPages: len(tags)/options.PerPage + 1,
	}

	return result, nil
}
