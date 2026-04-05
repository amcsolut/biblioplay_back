package profile

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// MaxAuthorCommunitySlugLen limite alinhado a profile_authors.slug e communities.slug.
const MaxAuthorCommunitySlugLen = 100

// SlugFromPenName gera identificador em minúsculas, números e hífens a partir do nome público do autor.
func SlugFromPenName(penName string) string {
	s := strings.TrimSpace(penName)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)
	s = strings.ToLower(s)

	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevHyphen = false
		case unicode.IsSpace(r) || r == '-' || r == '_':
			if b.Len() > 0 && !prevHyphen {
				b.WriteRune('-')
				prevHyphen = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "autor"
	}
	if len(out) > MaxAuthorCommunitySlugLen {
		out = strings.TrimRight(out[:MaxAuthorCommunitySlugLen], "-")
	}
	return out
}
