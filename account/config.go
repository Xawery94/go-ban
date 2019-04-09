package account

import "golang.org/x/text/language"

// DefaultLocale to be used when no locale is specified
var DefaultLocale = language.English.String()

type config struct {
	locales map[string]struct{}
}

// Opt describes account service configuration option
type Opt func(c *config)

// Locales configures set of allowed locales
func Locales(locales ...language.Tag) Opt {
	return func(c *config) {
		for _, locale := range locales {
			c.locales[locale.String()] = struct{}{}
		}
	}
}
