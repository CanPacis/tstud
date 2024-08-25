package cli

import (
	"github.com/CanPacis/tag-studio-core/controllers"
	"github.com/CanPacis/tag-studio-core/db"
	"github.com/alecthomas/kong"
)

var FileController *controllers.FileController
var TagController *controllers.TagController

func init() {
	dbs, err := db.Connect()
	if err != nil {
		panic(err)
	}
	FileController = controllers.NewFileController(dbs)
	TagController = controllers.NewTagController(dbs)
}

/*
tstud file index <filepath>
tstud file index --dir <dirpath>
tstud file unindex <filepath|dirpath>
tstud file tag <file id> <tag id>
tstud file untag <file id> <tag id>
tstud file meta set author <file id> <author>
tstud file meta unset author <file id>
tstud file list --page <page> --per-page <per page>
tstud file search term --tags <tag name list>
tstud file details <file id>

tstud tag create <tag name>
tstud tag delete <tag id>
tstud tag alias <tag id> <alias name>
tstud tag unalias <tag id> <alias id>
tstud tag parent <tag id> <parent tag id>
tstud tag list --page <page> --per-page <per page> [--all | --parent <parent id>]
tstud tag search term
*/

type Context struct {
	Debug bool
}

var cli struct {
	Debug bool `help:"Enable debug mode."`

	File struct {
		Index   FileIndexCmd   `cmd:"" help:"Index files and directories."`
		Unindex FileUnindexCmd `cmd:"" help:"Unindex files and directories."`
		Tag     FileTagCmd     `cmd:"" help:"Tag a file."`
		Untag   FileUntagCmd   `cmd:"" help:"Untag a file."`
		Details FileDetailsCmd `cmd:"" help:"See a file's details."`
		Search  FileSearchCmd  `cmd:"" help:"Search through files."`
		List    FileListCmd    `cmd:"" help:"List indexed files."`

		Meta struct {
			Set   FileMetaSetCmd   `cmd:""`
			Unset FileMetaUnsetCmd `cmd:""`
		} `cmd:""`
	} `cmd:"" help:"Work with files. Index, tag, list and search files."`

	Tag struct {
		Create  TagCreateCmd  `cmd:"" help:"Create a new tag."`
		Delete  TagDeleteCmd  `cmd:"" help:"Delete existing tag."`
		Alias   TagAliasCmd   `cmd:"" help:"Create an alias for a tag."`
		Unalias TagUnaliasCmd `cmd:"" help:"Remove an alias from a tag."`
		List    TagListCmd    `cmd:"" help:"List created tags."`
		Search  TagSearchCmd  `cmd:"" help:"Search through created tags."`
	} `cmd:"" help:"Work with tags. Create, delete and alias tags"`
}

func Run() {
	ctx := kong.Parse(&cli)
	err := ctx.Run(&Context{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}
