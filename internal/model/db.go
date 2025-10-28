package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID          uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	LineID      string    `bun:"line_id,notnull,unique"`
	DisplayName string    `bun:"display_name,nullzero"`
	CreatedAt   time.Time `bun:"created_at,default:current_timestamp"`
	UpdatedAt   time.Time `bun:"updated_at,default:current_timestamp"`

	Places []*Place `bun:"rel:has-many,join:id=added_by"`
	Votes  []*Vote  `bun:"rel:has-many,join:id=user_id"`
}

type Place struct {
	bun.BaseModel `bun:"table:places,alias:p"`

	ID        uuid.UUID     `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name      string        `bun:"name,notnull"`
	URL       string        `bun:"url,notnull,unique"`
	AddedBy   uuid.NullUUID `bun:"added_by,nullzero,type:uuid"`
	CreatedAt time.Time     `bun:"created_at,default:current_timestamp"`
	UpdatedAt time.Time     `bun:"updated_at,default:current_timestamp"`

	Author *User   `bun:"rel:belongs-to,join:added_by=id"`
	Votes  []*Vote `bun:"rel:has-many,join:id=place_id"`
	Tags   []*Tag  `bun:"m2m:place_tags,join:Place=Tag"`
}

type Vote struct {
	bun.BaseModel `bun:"table:votes,alias:v"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	PlaceID   uuid.UUID `bun:"place_id,type:uuid,notnull"`
	UserID    uuid.UUID `bun:"user_id,type:uuid,notnull"`
	Value     int16     `bun:"value,notnull"` // -1, 0, 1
	CreatedAt time.Time `bun:"created_at,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,default:current_timestamp"`

	User  *User  `bun:"rel:belongs-to,join:user_id=id"`
	Place *Place `bun:"rel:belongs-to,join:place_id=id"`
}

type Tag struct {
	bun.BaseModel `bun:"table:tags,alias:t"`

	ID   uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name string    `bun:"name,notnull,unique"`

	Places []*Place `bun:"m2m:place_tags,join:Tag=Place"`
}

type PlaceTag struct {
	bun.BaseModel `bun:"table:place_tags,alias:pt"`

	PlaceID uuid.UUID `bun:"place_id,pk,type:uuid"`
	TagID   uuid.UUID `bun:"tag_id,pk,type:uuid"`

	Place *Place `bun:"rel:belongs-to,join:place_id=id"`
	Tag   *Tag   `bun:"rel:belongs-to,join:tag_id=id"`
}

type PlaceSummary struct {
	bun.BaseModel `bun:"table:place_summary,alias:ps"`

	ID         uuid.UUID `bun:"id,type:uuid"`
	Name       string    `bun:"name"`
	URL        string    `bun:"url"`
	Score      float64   `bun:"score"`
	TotalVotes int       `bun:"total_votes"`
	Tags       []string  `bun:"tags,array"`
}
