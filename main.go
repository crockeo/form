package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

type AppContext struct {
	db              *gorm.DB
	jwtSigningToken []byte
}

func main() {
	if err := mainImpl(); err != nil {
		log.Fatal(err)
	}
}

func mainImpl() error {
	if err := loadEnv(); err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open("form.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&Form{}, &Response{}); err != nil {
		return err
	}

	jwtSigningToken := os.Getenv("JWT_SIGNING_TOKEN")
	if jwtSigningToken == "" {
		return fmt.Errorf("JWT_SIGNING_TOKEN missing")
	}

	ctx := &AppContext{db: db, jwtSigningToken: []byte(jwtSigningToken)}

	engine := gin.Default()
	engine.Use(identityMiddleware(ctx))

	engine.POST("/api/v1/form", createForm(ctx))
	engine.GET("/api/v1/form/:id", getForm(ctx))
	engine.POST("/api/v1/form/:id/response", postResponse(ctx))

	log.Print("Listening on 127.0.0.1:8000")
	engine.Run("127.0.0.1:8000")
	return nil
}

func loadEnv() error {
	env, err := os.ReadFile(".env")
	if err != nil {
		return err
	}
	for line := range strings.Lines(string(env)) {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, "=", 2)
		key := parts[0]
		value := parts[1]
		os.Setenv(key, value)
	}
	return nil
}

type CreateFormRequest struct {
	Prompt string `json:"prompt"`
}

type CreateFormResponse struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
}

func identityMiddleware(ctx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := getOrCreateIdentity(c, ctx.jwtSigningToken)
		if err != nil {
			c.Error(err)
			return
		}
		c.Set("identity", identity)
		c.Next()
	}
}

func createForm(ctx *AppContext) gin.HandlerFunc {
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
		if err := gorm.G[Form](ctx.db).Create(c.Request.Context(), &form); err != nil {
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

func getForm(ctx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Params[0].Value
		form, err := gorm.G[Form](ctx.db).Where("id = ?", id).Preload("Responses", func(db gorm.PreloadBuilder) error {
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

func postResponse(ctx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := PostResponseRequest{}
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		id := c.Params[0].Value
		form, err := gorm.G[Form](ctx.db).Where("id = ?", id).Preload("Responses", func(db gorm.PreloadBuilder) error {
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

		if err := gorm.G[Response](ctx.db).Create(c.Request.Context(), &newResponse); err != nil {
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

func getOrCreateIdentity(c *gin.Context, jwtSigningToken []byte) (string, error) {
	identityRaw, err := c.Cookie("identity")
	if errors.Is(err, http.ErrNoCookie) {
		identity, signedString, err := newIdentityToken(jwtSigningToken)
		if err != nil {
			return "", err
		}
		c.SetCookie("identity", signedString, 0, "/", "localhost", true, true)
		return identity, nil
	} else if err != nil {
		return "", err
	}

	token, err := jwt.Parse(
		identityRaw,
		func(token *jwt.Token) (any, error) {
			return jwtSigningToken, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return "", err
	}

	audience, err := token.Claims.GetAudience()
	if err != nil {
		return "", err
	}
	return audience[0], nil
}

func newIdentityToken(jwtSigningToken []byte) (string, string, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return "", "", err
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Audience: []string{uuid.String()},
		IssuedAt: jwt.NewNumericDate(now),
		Issuer:   "form",
	})
	signedString, err := token.SignedString(jwtSigningToken)
	if err != nil {
		return "", "", err
	}
	return uuid.String(), signedString, nil
}
