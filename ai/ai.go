package ai

import (
	"math/rand"
	"strings"
)

var (
	emojis = strings.Split(
		":thinking_face: :face_with_raised_eyebrow: :anguished: :exploding_head: :face_with_monocle: :see_no_evil: :hear_no_evil: :speak_no_evil:",
		" ",
	)
)

func Answer(text string) string {
	return text + "? " + emojis[rand.Intn(len(emojis))]
}
