package provider

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// acctestResourceName returns a prefix plus a short random suffix so parallel
// acceptance test runs don't collide on MISP's unique-name constraints.
func acctestResourceName(prefix string) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b strings.Builder
	for range 8 {
		b.WriteByte(letters[rand.IntN(len(letters))])
	}
	return fmt.Sprintf("%s-%s", prefix, b.String())
}
