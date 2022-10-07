package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"libman.mc/internal/validator"
)

type Book struct {
	ID			int64			`json:"id"`
	CreatedAt 	time.Time		`json:"-"`
	Title		string			`json:"title"`
	Authors 	[]string		`json:"authors"`
	Year 		int32			`json:"year,omitempty"`
	Publisher 	string			`json:"publisher"`
	Language	string			`json:"language,omitempty"`
	Pages 		int32			`json:"pages,omitempty"`
	Version 	int32			`json:"version"`
}

type BookModel struct {
	DB *sql.DB
}

func (m BookModel) Insert(book *Book) error{
	query := `
	INSERT INTO books (title, authors, year, publisher, language, pages)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, version`
	
	args := []interface{}{book.Title, pq.Array(book.Authors), book.Year, book.Publisher, book.Language, book.Pages}
	
	return m.DB.QueryRow(query, args...).Scan(&book.ID, &book.CreatedAt, &book.Version)
}

func (m BookModel) Get(id int64) (*Book, error){
	if id < 1{
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, authors, year, publisher, language, pages, version
		FROM books
		WHERE id = $1`
	
	var book Book
	err := m.DB.QueryRow(query, id).Scan(
		&book.ID,
		&book.CreatedAt,
		&book.Title,
		pq.Array(&book.Authors),
		&book.Year,
		&book.Publisher,
		&book.Language,
		&book.Pages,
		&book.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &book, nil
}

func (m BookModel) Update(book *Book) error{
	query := `
		UPDATE books
		SET title = $1, authors = $2, year = $3, publisher = $4, language = $5, pages = $6, version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`
	
	args := []interface{}{
		book.Title, 
		pq.Array(book.Authors), 
		book.Year, 
		book.Publisher, 
		book.Language, 
		book.Pages, 
		book.ID, 
		book.Version,
	}

	err := m.DB.QueryRow(query, args...).Scan(&book.Version)
	if err != nil {
		switch{
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflic
		default:
			return err
		}
	}

	return nil
}

func (m BookModel) Delete(id int64) error{
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM books
		WHERE id = $1`
	
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateBook(v *validator.Validator, book *Book) {
	v.Check(book.Title != "", "title", "must be provided")
	v.Check(len(book.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(book.Authors != nil, "authors", "must be provided")
	v.Check(len(book.Authors) >= 1, "authors", "must contain at least 1 author")
	v.Check(len(book.Authors) <= 5, "authors", "must not contain more than 5 authors")

	v.Check(book.Year != 0, "year", "must be provided")
	v.Check(book.Year >= 0, "year", "must be greater than 0")
	v.Check(book.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(book.Publisher != "", "publisher", "must be provided")
	v.Check(len(book.Publisher) <= 500, "publisher", "must not be more than 500 bytes long")

	v.Check(book.Language != "", "language", "must be provided")
	v.Check(len(book.Language) <= 500, "language", "must not be more than 500 bytes long")

	v.Check(book.Pages != 0, "pages", "must be provided")
	v.Check(book.Pages > 0, "pages", "must be a positive integer")
	
	v.Check(validator.Unique(book.Authors), "authors", "must not contain duplicate values")
}
