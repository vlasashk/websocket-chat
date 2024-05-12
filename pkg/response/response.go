package response

import (
	"fmt"
)

type Msg struct {
	UserID   int    `json:"user_id,omitempty"`
	Username string `json:"username"`
	Text     string `json:"text"`
}

type RegisterReq struct {
	Username string `json:"username"`
}

type RegisterResp struct {
	UserID int `json:"user_id"`
}

func (m Msg) Print() {
	fmt.Printf("<%s>:%s\n", m.Username, m.Text)
}
