package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CanPacis/tag-studio-core/controllers"
	"github.com/CanPacis/tag-studio-core/db"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
	"gorm.io/gorm"
)

func fileDtoToRows(items []any) []table.Row {
	result := []table.Row{}

	for _, item := range items {
		file, ok := item.(db.FileDTO)
		if !ok {
			continue
		}
		result = append(result, table.Row{fmt.Sprintf("%d", file.ID), file.Name, file.MimeType, filepath.Dir(file.FilePath)})
	}

	return result
}

func printFileDtoTable(title string, resource controllers.PaginatedResource) {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))

	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "File Name", Width: 28},
		{Title: "Mime Type", Width: 32},
		{Title: "Dir", Width: w - 76},
	}

	width := 0
	for _, column := range columns {
		width += column.Width
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Width(width).Align(lipgloss.Center)
	fmt.Println(titleStyle.Render(fmt.Sprintf("%s\n", title)))

	rows := fileDtoToRows(resource.Items)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(len(resource.Items)+1),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	s.Selected = s.Selected.Foreground(lipgloss.Color("f"))
	t.SetStyles(s)
	fmt.Println(t.View())

	pages := lipgloss.NewStyle().Bold(true).Faint(true).Width(width).Align(lipgloss.Center)
	fmt.Println(pages.Render(fmt.Sprintf("\n%d of %d", resource.Page+1, resource.TotalPages)))
}

func printFileDtoDetails(file db.FileDTO) {
	faint := lipgloss.NewStyle().Faint(true)
	bold := lipgloss.NewStyle().Bold(true)
	fmt.Println(faint.Render(fmt.Sprintf("File: %d", file.ID)))
	border := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	fmt.Println(border.Render(fmt.Sprintf("%s %s", bold.Render(file.Name), faint.Render(file.FilePath))))
	if len(file.Author) > 0 {
		fmt.Print(faint.Render("Author: "))
		fmt.Println(file.Author)
	}
	if len(file.Description) > 0 {
		fmt.Print(faint.Render("Description: "))
		fmt.Println(file.Description)
	}

	fmt.Print("\n")
	fmt.Println("Tags")
	box := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())
	tags := []string{}
	for _, tag := range file.Tags {
		tags = append(tags, box.Render(fmt.Sprintf(" %s ", tag.Name)))
	}

	if len(file.Tags) == 0 {
		fmt.Println(faint.Render("No tags"))
	} else {
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, tags...))
	}
}

type FileIndexCmd struct {
	Dir       bool     `short:"d" help:"Index a whole directory"`
	Recursive bool     `short:"r" help:"Index the directory recursively"`
	Exclude   []string `short:"e" help:"Exclude directory names"`

	Path string `arg:"" name:"path" help:"Paths to index." type:"path"`
}

func (c *FileIndexCmd) Run(ctx *Context) error {
	result, err := FileController.Index(c.Path, c.Recursive, c.Exclude)

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("resource already indexed")
		}
		return err
	}
	if len(result.Items) > 0 {
		printFileDtoTable("Indexed Files", *result)
	}
	return nil
}

type FileUnindexCmd struct {
	Dir       bool     `short:"d" help:"Index a whole directory"`
	Recursive bool     `short:"r" help:"Index the directory recursively"`
	Exclude   []string `short:"e" help:"Exclude directory names"`

	Path string `arg:"" name:"path" help:"Paths to index." type:"path"`
}

func (c *FileUnindexCmd) Run(ctx *Context) error {
	result, err := FileController.Unindex(c.Path, c.Recursive, c.Exclude)
	if err != nil {
		return err
	}
	if len(result.Items) > 0 {
		printFileDtoTable("Deindexed Files", *result)
	}
	return nil
}

type FileListCmd struct {
	Page    int `cmd:"" short:"p" help:"Select list page."`
	PerPage int `cmd:"" short:"l" help:"Selec items per page."`
}

func (c *FileListCmd) Run(ctx *Context) error {
	var result *controllers.PaginatedResource
	page := c.Page - 1

	if page < 0 {
		page = 0
	}

	if c.PerPage == 0 {
		c.PerPage = 12
	}

	var err error
	result, err = FileController.List(controllers.ListOptions{
		PerPage: c.PerPage,
		Page:    page,
	})

	if err != nil {
		return err
	}

	printFileDtoTable("List Result", *result)
	return nil
}

type FileTagCmd struct {
	File uint `arg:"" name:"file id" help:"File id to tag"`
	Tag  uint `arg:"" name:"tag id" help:"Tag id to attach"`
}

func (c *FileTagCmd) Run(ctx *Context) error {
	file, tag, err := FileController.Tag(c.File, c.Tag)
	if err != nil {
		return err
	}

	fmt.Printf("Tagged file '%s' with '%s'\n", file.Name, tag.Name)
	return nil
}

type FileUntagCmd struct {
	File uint `arg:"" name:"file id" help:"File id to remove the tag"`
	Tag  uint `arg:"" name:"tag id" help:"Tag id to detach"`
}

func (c *FileUntagCmd) Run(ctx *Context) error {
	file, tag, err := FileController.Untag(c.File, c.Tag)
	if err != nil {
		return err
	}

	fmt.Printf("Removed tag '%s' from file '%s'\n", tag.Name, file.Name)
	return nil
}

type FileDetailsCmd struct {
	ID   uint   `cmd:"" short:"i" help:"Find a file with its id."`
	Path string `cmd:"" short:"p" help:"Find a file with its path." type:"path"`
}

func (c *FileDetailsCmd) Run(ctx *Context) error {
	var file *db.FileDTO
	var err error
	if c.ID != 0 {
		file, err = FileController.FindByID(c.ID)
	} else {
		file, err = FileController.FindByPath(c.Path)
	}

	if err != nil {
		return err
	}

	printFileDtoDetails(*file)
	return nil
}

type FileMetaSetCmd struct {
	Author struct {
		FileID uint   `arg:"" name:"file id" help:"File id to attach an author to."`
		Value  string `arg:"" name:"author" help:"File author name."`
	} `cmd:"" help:"Set the author meta field of a file."`
	Description struct {
		FileID uint   `arg:"" name:"file id" help:"File id to attach a description to."`
		Value  string `arg:"" name:"description" help:"File description value."`
	} `cmd:"" help:"Set the description meta field of a file."`
}

func (c *FileMetaSetCmd) Run(ctx *Context) error {
	var fileId uint
	if c.Author.FileID != 0 {
		fileId = c.Author.FileID
	} else if c.Description.FileID != 0 {
		fileId = c.Description.FileID
	}

	meta := controllers.FileMetaData{}

	if len(c.Author.Value) > 0 {
		meta.Author = &c.Author.Value
	}
	if len(c.Description.Value) > 0 {
		meta.Description = &c.Description.Value
	}

	file, err := FileController.SetMeta(fileId, meta)
	if err != nil {
		return err
	}
	fmt.Print("Updated metadata of the file\n\n")
	printFileDtoDetails(*file)
	return nil
}

type FileMetaUnsetCmd struct {
	Author struct {
		FileID uint `arg:"" name:"file id" help:"File id to remove the author from."`
	} `cmd:"" help:"Remove the author meta field of a file."`
	Description struct {
		FileID uint `arg:"" name:"file id" help:"File id to remove the description from."`
	} `cmd:"" help:"Remove the description meta field of a file."`
}

func (c *FileMetaUnsetCmd) Run(ctx *Context) error {
	var fileId uint
	meta := controllers.FileMetaData{}
	empty := ""

	if c.Author.FileID != 0 {
		fileId = c.Author.FileID
		meta.Author = &empty
	} else if c.Description.FileID != 0 {
		fileId = c.Description.FileID
		meta.Description = &empty
	}

	file, err := FileController.SetMeta(fileId, meta)
	if err != nil {
		return err
	}
	fmt.Print("Updated metadata of the file\n\n")
	printFileDtoDetails(*file)
	return nil
}

type FileSearchCmd struct {
	Term string `arg:"" help:"Tags to search"`

	Tags        []string `cmd:"" short:"t" help:"Tags to search."`
	Author      string   `cmd:"" short:"a" help:"Search author of a file."`
	Description string   `cmd:"" short:"d" help:"Search description of a file."`
}

func (c *FileSearchCmd) Run(ctx *Context) error {
	options := controllers.SearchOptions{
		Term:        c.Term,
		Tags:        c.Tags,
		Author:      c.Author,
		Description: c.Description,
		ListOptions: controllers.ListOptions{
			Page:    0,
			PerPage: 10,
		},
	}
	result, err := FileController.Search(options)
	if err != nil {
		return err
	}

	for _, file := range result.Items {
		fmt.Println(file)
	}
	fmt.Printf("\n%d of %d\n", result.Page+1, result.TotalPages)

	return nil
}
