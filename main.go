package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Form struct {
	gorm.Model

	ID        string `gorm:"primaryKey"`
	Prompt    string
	Responses []Response `gorm:"foreignKey:FormID"`
}

type Response struct {
	gorm.Model

	ID     string `gorm:"primaryKey"`
	FormID string
	Text   string
}

func main() {
	if err := mainImpl(); err != nil {
		log.Fatal(err)
	}
}

func mainImpl() error {
	db, err := gorm.Open(sqlite.Open("form.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&Form{}, &Response{}); err != nil {
		return err
	}

	engine := gin.Default()

	engine.POST("/api/v1/form", createForm(db))
	engine.GET("/api/v1/form/:id", getForm(db))
	engine.POST("/api/v1/form/:id/response", postResponse(db))

	log.Print("Listening on 127.0.0.1:8000")
	engine.Run("127.0.0.1:8000")
	return nil
}

type CreateFormRequest struct {
	Prompt string `json:"prompt"`
}

type CreateFormResponse struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
}

func createForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := CreateFormRequest{}
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		id, err := uuid.NewV7()
		if err != nil {
			c.Error(err)
			return
		}

		form := Form{
			ID:     id.String(),
			Prompt: req.Prompt,
		}
		if err := gorm.G[Form](db).Create(c.Request.Context(), &form); err != nil {
			c.Error(err)
			return
		}

		c.JSON(200, CreateFormResponse{ID: form.ID, Prompt: form.Prompt})
	}
}

type GetFormRequest struct {
	ID string `json:"id"`
}

type FormResponse struct {
	ID        string              `json:"id"`
	Prompt    string              `json:"prompt"`
	Responses []FormResponseEntry `json:"responses"`
}

type FormResponseEntry struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func getForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Params[0].Value
		form, err := gorm.G[Form](db).Where("id = ?", id).Preload("Responses", func(db gorm.PreloadBuilder) error {
			return nil
		}).First(c.Request.Context())
		if err != nil {
			c.Error(err)
			return
		}

		responses := make([]FormResponseEntry, len(form.Responses))
		for i, response := range form.Responses {
			responses[i] = FormResponseEntry{
				ID:   response.ID,
				Text: response.Text,
			}
		}

		c.JSON(
			200,
			FormResponse{
				ID:        form.ID,
				Prompt:    form.Prompt,
				Responses: responses,
			},
		)
	}
}

type PostResponseRequest struct {
	Text string `json:"text"`
}

func postResponse(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := PostResponseRequest{}
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		id := c.Params[0].Value
		form, err := gorm.G[Form](db).Where("id = ?", id).Preload("Responses", func(db gorm.PreloadBuilder) error {
			return nil
		}).First(c.Request.Context())
		if err != nil {
			c.Error(err)
			return
		}

		uuid, err := uuid.NewV7()
		if err != nil {
			c.Error(err)
			return
		}

		newResponse := Response{
			ID:     uuid.String(),
			FormID: id,
			Text:   req.Text,
		}

		if err := gorm.G[Response](db).Create(c.Request.Context(), &newResponse); err != nil {
			c.Error(err)
			return
		}

		form.Responses = append(form.Responses, newResponse)
		responses := make([]FormResponseEntry, len(form.Responses))
		for i, response := range form.Responses {
			responses[i] = FormResponseEntry{
				ID:   response.ID,
				Text: response.Text,
			}
		}

		c.JSON(
			200,
			FormResponse{
				ID:        form.ID,
				Prompt:    form.Prompt,
				Responses: responses,
			},
		)
	}
}
