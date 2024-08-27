package proto

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/CanPacis/tstud-core/controllers"
	"github.com/CanPacis/tstud-core/p2pjson"
)

func CreateTag(r *p2pjson.Request) *p2pjson.Response {
	var data struct {
		Name     string `json:"name"`
		ParentID int    `json:"parent_id"`
	}
	err := json.Unmarshal(r.Get("body").([]byte), &data)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusBadRequest, err)
	}

	var parentId *int

	if data.ParentID != 0 {
		parentId = &data.ParentID
	}
	tag, err := TagController.Create(data.Name, parentId)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	encoded, err := json.Marshal(tag)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	return p2pjson.NewResponse(r, p2pjson.StatusCreated, bytes.NewBuffer(encoded))
}

func DeleteTag(r *p2pjson.Request) *p2pjson.Response {
	var data struct {
		ID uint `json:"id"`
	}
	err := json.Unmarshal(r.Get("body").([]byte), &data)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusBadRequest, err)
	}

	tag, err := TagController.Delete(data.ID)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	encoded, err := json.Marshal(tag)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	return p2pjson.NewResponse(r, p2pjson.StatusOK, bytes.NewBuffer(encoded))
}

func AliasTag(r *p2pjson.Request) *p2pjson.Response {
	var data struct {
		ID   uint   `json:"id"`
		Name string `json:"string"`
	}
	err := json.Unmarshal(r.Get("body").([]byte), &data)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusBadRequest, err)
	}

	err = TagController.Alias(data.ID, data.Name)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	encoded, _ := json.Marshal(map[string]any{"message": "done"})
	return p2pjson.NewResponse(r, p2pjson.StatusOK, bytes.NewBuffer(encoded))
}

func UnaliasTag(r *p2pjson.Request) *p2pjson.Response {
	var data struct {
		ID   uint   `json:"id"`
		Name string `json:"string"`
	}
	err := json.Unmarshal(r.Get("body").([]byte), &data)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusBadRequest, err)
	}

	err = TagController.Unlias(data.ID, data.Name)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	encoded, _ := json.Marshal(map[string]any{"message": "done"})
	return p2pjson.NewResponse(r, p2pjson.StatusOK, bytes.NewBuffer(encoded))
}

func ParentTag(r *p2pjson.Request) *p2pjson.Response {
	return p2pjson.ErrorResponse(r, p2pjson.StatusNotImplemented, errors.New("tag/parent not implemented"))
}

func ListTag(r *p2pjson.Request) *p2pjson.Response {
	var data struct {
		Page     int  `json:"page"`
		PerPage  int  `json:"per_page"`
		ParentID int  `json:"parent_id"`
		All      bool `json:"all"`
	}
	err := json.Unmarshal(r.Get("body").([]byte), &data)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusBadRequest, err)
	}

	result, err := TagController.List(&data.ParentID)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	return p2pjson.NewResponse(r, p2pjson.StatusOK, bytes.NewBuffer(encoded))
}

func SearchTag(r *p2pjson.Request) *p2pjson.Response {
	var data struct {
		Page    int    `json:"page"`
		PerPage int    `json:"per_page"`
		Term    string `json:"term"`
	}
	err := json.Unmarshal(r.Get("body").([]byte), &data)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusBadRequest, err)
	}

	result, err := TagController.Search(data.Term, controllers.ListOptions{
		Page:    data.Page,
		PerPage: data.PerPage,
	})
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		return p2pjson.ErrorResponse(r, p2pjson.StatusInternalServerError, err)
	}

	return p2pjson.NewResponse(r, p2pjson.StatusOK, bytes.NewBuffer(encoded))
}
