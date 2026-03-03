package dialogue

import "time"

type Dialogue struct {
	ID        int64      `db:"id" json:"id"`
	User1ID   int64      `db:"user1_id" json:"user1_id"`
	User2ID   int64      `db:"user2_id" json:"user2_id"`
	IsActive  bool       `db:"is_active" json:"is_active"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	ClosedAt  *time.Time `db:"closed_at" json:"closed_at,omitempty"`
}
