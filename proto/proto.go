package proto

import (
	"io"

	"github.com/CanPacis/tstud-core/controllers"
	"github.com/CanPacis/tstud-core/db"
	"github.com/CanPacis/tstud-core/p2pjson"
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
/file/index { path: string; dir: boolean; recursive: boolean; exclude: string[]; }
/file/unindex { path: string; dir: boolean; recursive: boolean; exclude: string[]; }
/file/rename { oldpath: string; newpath: string; }
/file/tag { file_id: number; tag_id: number; }
/file/untag { file_id: number; tag_id: number; }
/file/meta/set/author { file_id:number; author: string; }
/file/meta/unset/author { file_id:number; }
/file/meta/set/description { file_id:number; description: string; }
/file/meta/unset/description { file_id:number; }
/file/list { page: number; per_page: number; }
/file/search { page: number; per_page: number; term: string; tags: string[] }
/file/details { file_id: number; }

/tag/create { name: string; parent_id: number; }
/tag/delete { tag_id: number; }
/tag/alias { tag_id: number; alias_name: string; }
/tag/unalias { tag_id: number; alias_id: number; }
/tag/parent { tag_id: number; parent_tag_id: number; }
/tag/list { page: number; per_page: number; parent_id: number; all: boolean; }
/tag/search { page: number; per_page: number; term: string }
*/

func Run() {
	mux := p2pjson.NewMux()
	peer := p2pjson.New(p2pjson.NewStdIOPeer())

	mux.HandleFunc("/tag/create", JsonMiddleWare(CreateTag))
	mux.HandleFunc("/tag/delete", JsonMiddleWare(DeleteTag))
	mux.HandleFunc("/tag/alias", JsonMiddleWare(AliasTag))
	mux.HandleFunc("/tag/unalias", JsonMiddleWare(UnaliasTag))
	mux.HandleFunc("/tag/parent", JsonMiddleWare(ParentTag))
	mux.HandleFunc("/tag/list", JsonMiddleWare(ListTag))
	mux.HandleFunc("/tag/searxh", JsonMiddleWare(SearchTag))

	peer.Listen(mux)
}

func JsonMiddleWare(next p2pjson.HandlerFunc) p2pjson.HandlerFunc {
	return func(r *p2pjson.Request) *p2pjson.Response {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			return p2pjson.ErrorResponse(r, p2pjson.StatusBadRequest, err)
		}
		r.Set("body", raw)

		return next(r)
	}
}
