package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (b *Bot) IsAdmin(userID int64) bool {
	adminIDsStr := os.Getenv("ADMIN_IDS")
	if adminIDsStr == "" {
		return false
	}

	adminIDs := strings.Split(adminIDsStr, ",")
	for _, idStr := range adminIDs {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err == nil && id == int64(userID) {
			return true
		}
	}

	return false
}

func (b *Bot) IsAdminGroup(groupID int64) bool {
	adminIDsStr := os.Getenv("ADMIN_GROUP_IDS")
	fmt.Println("adminIDsStr:", adminIDsStr, "groupID:", groupID)
	if adminIDsStr == "" {
		return false
	}

	adminIDs := strings.Split(adminIDsStr, ",")
	for _, idStr := range adminIDs {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err == nil && id == groupID {
			return true
		}
	}

	return false
}
