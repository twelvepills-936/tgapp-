package models

import "time"

type Profile struct {
    ID               int64
    Name             string
    TelegramID       string
    Avatar           string
    Location         string
    Role             string
    Description      string
    TelegramInitData string
    Username         string
    Verified         bool
    CreatedAt        time.Time
    UpdatedAt        time.Time
}


