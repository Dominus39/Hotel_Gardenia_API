package utils

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"fmt"
)

func TrackUserActivity(userID int, description string) {
	activityLog := entity.UserHistory{
		UserID:      userID,
		Description: description,
	}
	if err := config.DB.Create(&activityLog).Error; err != nil {
		fmt.Println("Error logging activity:", err)
	}
}
