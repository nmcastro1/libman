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
