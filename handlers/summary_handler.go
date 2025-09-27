package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log"
	// REMOVED: "go.mongodb.org/mongo-driver/bson" - No longer needed.
	"net/http"
	"time"
	"ultra-chat-backend/models" // ADDED: To use the new SummaryItem model.
	"ultra-chat-backend/repositories"
	"ultra-chat-backend/utils"
)

type SummaryHandler struct {
	// CHANGED: We now depend on the interface, not the concrete implementation.
	// This makes the handler completely decoupled from the database.
	repo repositories.SummaryRepository
}

// NewSummaryHandler now accepts the interface.
func NewSummaryHandler(summaryRepo repositories.SummaryRepository) *SummaryHandler {
	return &SummaryHandler{repo: summaryRepo}
}

func (h *SummaryHandler) CreateSummary(c echo.Context) error {
	type RequestBody struct {
		Content   string `json:"content"`
		ServerID  string `json:"server_id"`
		IsPrivate bool   `json:"is_private"`
		UserID    string `json:"user_id"`
	}

	var body RequestBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if body.Content == "" || body.ServerID == "" || body.UserID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
	}

	// This method call remains the same as the interface was preserved.
	exists, dbErr := h.repo.CheckUserExists(body.UserID)
	if dbErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	if !exists {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: User not found"})
	}

	summaryID := uuid.New().String()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	// CHANGED: We now construct a proper model struct instead of passing loose parameters.
	// This is cleaner and safer.
	newSummary := models.SummaryItem{
		UserID:    body.UserID,
		SummaryID: summaryID,
		ServerID:  body.ServerID,
		IsPrivate: body.IsPrivate,
		Content:   body.Content,
		CreatedAt: createdAt,
		UpdatedAt: createdAt, // On creation, created_at and updated_at are the same.
	}

	err := h.repo.AddSummary(newSummary)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create summary"})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message":    "Summary created successfully",
		"summary_id": summaryID,
	})
}

func (h *SummaryHandler) GetSummaries(c echo.Context) error {
	userID := c.Request().Header.Get("ID")
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// CHANGED: Removed the bson.M filter and now call the specific method.
	// This call directly translates to querying the DynamoDB table by its partition key.
	summaries, err := h.repo.GetSummariesByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve summaries"})
	}

	// If no summaries are found, return an empty list, not an error.
	if summaries == nil {
		summaries = []models.SummaryItem{}
	}

	return c.JSON(http.StatusOK, summaries)
}

func (h *SummaryHandler) UpdateSummary(c echo.Context) error {
	type RequestBody struct {
		SummaryID string `json:"summary_id"`
		ServerID  string `json:"server_id"`  // Kept in body for client consistency, but not used in the repo call
		IsPrivate bool   `json:"is_private"` // Kept in body for client consistency, but not used in the repo call
		Content   string `json:"content"`
	}

	var body RequestBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	userID := c.Request().Header.Get("ID")
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// The body must contain the specific summary_id to update.
	if body.SummaryID == "" || body.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields: summary_id and content"})
	}

	// CHANGED: We now call the more precise UpdateSummaryContent method.
	// This method directly updates a specific summary item using its full primary key,
	// which is much more reliable than the old query.
	err := h.repo.UpdateSummaryContent(userID, body.SummaryID, body.Content)
	if err != nil {
		// The repository now returns a specific error for "not found".
		if err.Error() == "no matching summary found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update summary"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Summary updated successfully"})
}

func (h *SummaryHandler) DeleteSummary(c echo.Context) error {
	type RequestBody struct {
		SummaryID string `json:"summary_id"`
	}
	var body RequestBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	userID := c.Request().Header.Get("ID")
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// <<< --- THIS IS THE CRITICAL DEBUG LINE --- >>>
	log.Printf("Received delete request for UserID: '%s' with SummaryID from body: '%s'", userID, body.SummaryID)

	if err := h.repo.DeleteSummary(userID, body.SummaryID); err != nil {
		if err.Error() == "no matching summary found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete summary"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Summary deleted successfully"})
}

// This function does not interact with the repository, so it remains unchanged.
func (h *SummaryHandler) IsAuthenticated(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	fmt.Println(authHeader)
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Missing token"})
	}

	accessToken := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	} else {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Invalid token format"})
	}

	userInfo, err := utils.FetchUserInfo(accessToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Failed to validate token"})
	}

	userID, ok := userInfo["id"].(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Invalid user data"})
	}

	c.Set("userID", userID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "Authenticated",
		"user_id":   userID,
		"user_info": userInfo,
	})
}
