package utils

import "github.com/vlasashk/websocket-chat/pkg/response"

func FlipMessageOrder(messages []response.Msg) {
	end := len(messages) - 1
	for start := 0; start < len(messages)/2; start, end = start+1, end-1 {
		messages[start], messages[end] = messages[end], messages[start]
	}
}
