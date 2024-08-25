package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CanPacis/tag-studio-core/controllers"
	"github.com/CanPacis/tag-studio-core/db"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"gorm.io/gorm"
)

type TagCreateCmd struct {
	Parent int `cmd:"" short:"p" name:"parent" help:"Set a parent for this tag."`

	Name string `arg:"" name:"name" help:"Tag name to create."`
}

func (c *TagCreateCmd) Run(ctx *Context) error {
	var parent *int

	if c.Parent != 0 {
		parent = &c.Parent
	}

	tag, err := TagController.Create(c.Name, parent)
	fmt.Printf("Created tag %s\n", tag.Name)
	return err
}

type TagDeleteCmd struct {
	ID uint `arg:"" name:"id" help:"Tag id to delete."`
}

func (c *TagDeleteCmd) Run(ctx *Context) error {
	tag, err := TagController.Delete(c.ID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("tag with id %d not found", c.ID)
	}

	fmt.Printf("Deleted tag %s\n", tag.Name)
	return err
}

type TagAliasCmd struct {
	Tag   uint   `arg:"" name:"tag id" help:"Target tag id."`
	Alias string `arg:"" name:"alias" help:"New alias for the tag."`
}

func (c *TagAliasCmd) Run(ctx *Context) error {
	return TagController.Alias(c.Tag, c.Alias)
}

type TagUnaliasCmd struct {
	Tag   uint   `arg:"" name:"tag id" help:"Target tag id."`
	Alias string `arg:"" name:"alias" help:"Alias to be removed from the tag."`
}

func (c *TagUnaliasCmd) Run(ctx *Context) error {
	return TagController.Unlias(c.Tag, c.Alias)
}

func tagDtoToRows(items []any) []table.Row {
	result := []table.Row{}

	for _, item := range items {
		tag, ok := item.(db.TagDTO)
		if !ok {
			continue
		}
		aliases := []string{}
		for _, alias := range tag.Aliases {
			aliases = append(aliases, alias.Name)
		}
		if len(aliases) == 0 {
			aliases = append(aliases, "-")
		}
		var parentId string

		if tag.Parent != nil {
			parentId = fmt.Sprintf("%d", tag.Parent.ID)
		} else {
			parentId = "-"
		}

		result = append(result, table.Row{fmt.Sprintf("%d", tag.ID), parentId, tag.Name, strings.Join(aliases, ", ")})
	}
	return result
}

func printTagDtoTable(result controllers.PaginatedResource) {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Parent ID", Width: 12},
		{Title: "Tag Name", Width: 32},
		{Title: "Aliases", Width: 32},
	}

	rows := tagDtoToRows(result.Items)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(len(result.Items)+1),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	s.Selected = s.Selected.Foreground(lipgloss.Color("f"))
	t.SetStyles(s)
	fmt.Println(t.View())
}

type TagListCmd struct {
	Parent int  `cmd:"" short:"p" help:"List tags with specified parents.."`
	All    bool `cmd:"" short:"a" help:"List all tags. Cannot be used with -p"`
}

func (c *TagListCmd) Run(ctx *Context) error {
	var result *controllers.PaginatedResource
	var err error
	var parent *int

	if c.All {
		c.Parent = -1
		parent = &c.Parent
	} else {
		if c.Parent != 0 {
			parent = &c.Parent
		}
	}
	result, err = TagController.List(parent)

	if err != nil {
		return err
	}

	printTagDtoTable(*result)
	return nil
}

type TagSearchCmd struct {
	Term string `arg:"" help:"Search term."`
}

func (c *TagSearchCmd) Run(ctx *Context) error {
	result, err := TagController.Search(c.Term, controllers.ListOptions{
		Page:    0,
		PerPage: 12,
	})
	if err != nil {
		return err
	}

	printTagDtoTable(*result)
	return nil
}
