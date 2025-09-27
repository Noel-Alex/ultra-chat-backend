package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"ultra-chat-backend/models"
	"ultra-chat-backend/repositories"
	"ultra-chat-backend/utils"
)

type AuthHandler struct {
	repo repositories.UserRepository
}

func NewAuthHandler(repo repositories.UserRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

// Login method remains unchanged as it has no database interaction.
func (h *AuthHandler) Login(c echo.Context) error {
	clientID := os.Getenv("CLIENT_ID")
	redirectURI := os.Getenv("REDIRECT_URI")
	scope := os.Getenv("SCOPE")

	url := "https://discord.com/api/oauth2/authorize?client_id=" + clientID + "&redirect_uri=" + redirectURI + "&response_type=code&scope=" + scope
	return c.JSON(http.StatusOK, map[string]string{
		"url": url,
	})
}

func (h *AuthHandler) Callback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No code provided"})
	}

	tokens, err := utils.ExchangeCodeForTokens(code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	accessToken, ok := tokens["access_token"].(string)
	if !ok || accessToken == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid access token"})
	}

	userInfo, err := utils.FetchUserInfo(accessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	userID, ok := userInfo["id"].(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid user ID"})
	}

	existingUser, _ := h.repo.FindUserByID(userID)

	if existingUser != nil {
		// CHANGED: We replaced `bson.M` with a generic `map[string]interface{}`.
		// This decouples the handler from the database implementation. The repository
		// will be responsible for translating this map into a DynamoDB update expression.
		updateData := map[string]interface{}{
			"token":         tokens,
			"username":      userInfo["username"],
			"discriminator": userInfo["discriminator"],
		}
		if err := h.repo.UpdateUser(userID, updateData); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	} else {
		userUUID := uuid.New().String()
		newUser := &models.User{
			ID:            userID,
			UUID:          userUUID,
			Token:         tokens,
			Username:      userInfo["username"].(string),
			Discriminator: userInfo["discriminator"].(string),
		}
		if err := h.repo.CreateUser(newUser); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	return c.JSON(http.StatusOK, "Successfully Authenticated")
}

// Profile method remains unchanged as it has no database interaction.
func (h *AuthHandler) Profile(c echo.Context) error {
	// Get the token from the Authorization header
	token := c.Request().Header.Get("Authorization")
	fmt.Println("token", token)
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Fetch user information using the token
	userInfo, err := utils.FetchUserInfo(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// Respond with the user information
	return c.JSON(http.StatusOK, userInfo)
}
