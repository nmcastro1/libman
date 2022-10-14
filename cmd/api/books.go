package main

import (
	"errors"
	"fmt"
	"net/http"

	"libman.mc/internal/data"
	"libman.mc/internal/validator"
)

func (app *application) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title 		string		`json:title`
		Authors 	[]string 	`json:authors`
		Year 		int32 		`json:year`
		Publisher	string		`json:publisher`
		Language 	string		`json:language`
		Pages		int32		`json:pages`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	book := &data.Book{
		Title: 		input.Title,
		Authors: 	input.Authors,
		Year: 		input.Year,
		Publisher:  input.Publisher,
		Language:   input.Language,
		Pages:  	input.Pages,
	}

	v := validator.New()

	if data.ValidateBook(v, book); !v.Valid(){
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	
	err = app.models.Books.Insert(book)
	if err != nil{
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("location", fmt.Sprintf("/v1/books/%d", book.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"book": book}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"book": book}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)
	if err != nil {
		switch{
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title 		*string		`json:title`
		Authors 	[]string 	`json:authors`
		Year 		*int32 		`json:year`
		Publisher	*string		`json:publisher`
		Language 	*string		`json:language`
		Pages		*int32		`json:pages`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		book.Title = *input.Title
	}

	if input.Authors != nil {
		book.Authors = input.Authors
	}

	if input.Year != nil {
		book.Year = *input.Year
	}

	if input.Publisher != nil {
		book.Publisher = *input.Publisher
	}

	if input.Language != nil {
		book.Language = *input.Language
	}

	if input.Pages != nil {
		book.Pages = *input.Pages
	}
	
	v := validator.New()
	if data.ValidateBook(v, book); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Books.Update(book)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflic):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"book": book}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	err = app.models.Books.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "book successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title 		string
		Authors 	[]string
		Publisher 	string
		Language	string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Authors = app.readCSV(qs, "authors", []string{})
	input.Publisher = app.readString(qs, "publisher", "")
	input.Language = app.readString(qs, "language", "")
	
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 10, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "title", "year", "pages","-id", "-title", "-year", "-pages"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	books, err := app.models.Books.GetAll(input.Title, input.Authors, input.Publisher, input.Language, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"books": books}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
} 
