package api

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/account"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
)

// deleteMe closes the signed-in account. The caller must echo the current
// username in the `confirm` field, mirroring the typed confirmation in the
// HTML settings form, so an accidental or cross-site request cannot delete an
// account by itself.
func deleteMe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusUnauthorized, "Please sign in")
		return
	}

	var req struct {
		Confirm string `json:"confirm" form:"confirm"`
	}
	if err := c.ShouldBind(&req); err != nil {
		apiError(c, http.StatusBadRequest, "Send {\"confirm\":\"your-username\"}")
		return
	}
	if strings.TrimSpace(req.Confirm) != user.Username {
		apiError(c, http.StatusUnprocessableEntity, "Confirmation does not match the account username")
		return
	}

	err := account.New(database.DB).Delete(user.ID)
	switch err {
	case nil:
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Status(http.StatusNoContent)
	case account.ErrLastSuperuser:
		apiError(c, http.StatusConflict, "The last superuser cannot be deleted")
	case account.ErrNotFound:
		apiError(c, http.StatusNotFound, "Account not found")
	default:
		apiError(c, http.StatusInternalServerError, "Could not delete the account")
	}
}

// exportMe returns the signed-in user's own data as a JSON attachment.
func exportMe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusUnauthorized, "Please sign in")
		return
	}

	data, err := account.New(database.DB).Export(user.ID)
	if err != nil {
		if err == account.ErrNotFound {
			apiError(c, http.StatusNotFound, "Account not found")
			return
		}
		apiError(c, http.StatusInternalServerError, "Could not build the export")
		return
	}

	name := exportFilename(user.Username)
	c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, data)
}

// exportFilename keeps the download name to a safe ASCII subset; usernames may
// contain characters that are not valid in a Content-Disposition filename.
func exportFilename(username string) string {
	var b strings.Builder
	for _, r := range username {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		}
		if b.Len() >= 48 {
			break
		}
	}
	slug := b.String()
	if slug == "" {
		slug = "account"
	}
	return "sandwitches-export-" + slug + ".json"
}
