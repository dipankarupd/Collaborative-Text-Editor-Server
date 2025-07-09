package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dipankarupd/text-editor/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


func CreateDocument() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		authorIdVal, exist := ctx.Get("userid")

		if !exist {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}
		authorId, ok := authorIdVal.(uuid.UUID)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		} 
		
		doc := models.Document {
			ID: uuid.New(),
			AuthorID: authorId,
			Title: "Untitled Document",
			Content: json.RawMessage(`[]`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&doc).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Document"})
			return
		}

		var author models.User
		if err := db.First(&author, "id = ?", authorId).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch author name"})
			return
		}

		response := models.DocResponse{
			ID: doc.ID,
			Author: models.Author{
				ID: doc.AuthorID,
				Name: author.Name,
			},
			Title: doc.Title,
			Content: doc.Content,
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}

		ctx.JSON(http.StatusCreated, response)
	}
}

func GetUserDocuments() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorIdVal, exist := ctx.Get("userid")
		if !exist {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		authorNameVal, nameExist := ctx.Get("name")
		if !nameExist {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Author name not found"})
			return
		}

		authorId, ok := authorIdVal.(uuid.UUID)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		}

		authorName, ok := authorNameVal.(string)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid author name format"})
			return
		}

		var docs []models.Document
		if err := db.Where("author_id = ?", authorId).Find(&docs).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching the documents"})
			return
		}

		// Convert to []DocResponse
		docResponses := make([]models.DocResponse, len(docs))
		for i, d := range docs {
			docResponses[i] = models.DocResponse{
				ID: d.ID,
				Author: models.Author{
					ID:   d.AuthorID,
					Name: authorName,
				},
				Title:     d.Title,
				Content:   d.Content,
				CreatedAt: d.CreatedAt,
				UpdatedAt: d.UpdatedAt,
			}
		}

		ctx.JSON(http.StatusOK, docResponses)
	}
}

func GetDocumentByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		docId := ctx.Param("id")

		var doc models.Document

		if err := db.First(&doc, "id = ?", docId).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching the Document"})
			return
		}
		var author models.User
		if err := db.First(&author, "id = ?", doc.AuthorID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching the Author of the document"})
			return
		} 

		response := models.DocResponse {
			ID: doc.ID,
			Author: models.Author{
				ID: doc.AuthorID,
				Name: author.Name,
			},
			Title: doc.Title,
			Content: doc.Content,
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}
		ctx.JSON(http.StatusOK, response)
	}
}
func UpdateDocumentTitle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Parse document ID from URL
		docIDParam := ctx.Param("id")
		docID, err := uuid.Parse(docIDParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
			return
		}
		// Bind request body
		var body struct {
			Title string `json:"title"`
		}
		if err := ctx.ShouldBindJSON(&body); err != nil || body.Title == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
			return
		}
		// Check if document exists and belongs to the user
		authorIdVal, exist := ctx.Get("userid")
		if !exist {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}
		authorID, ok := authorIdVal.(uuid.UUID)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		}
		
		// First check if document exists at all
		var doc models.Document
		if err := db.First(&doc, "id = ?", docID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			return
		}
		
		// Then check if user owns the document
		if doc.AuthorID != authorID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this document"})
			return
		}
		
		// Update title and UpdatedAt
		doc.Title = body.Title
		doc.UpdatedAt = time.Now()
		if err := db.Save(&doc).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
			return
		}
		// Return response
		ctx.JSON(http.StatusOK, gin.H{
			"success":   "ok",
			"new_title": doc.Title,
		})
	}
}