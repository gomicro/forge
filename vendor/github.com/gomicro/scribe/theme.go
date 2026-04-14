package scribe

import "fmt"

type Theme struct {
	Describe func(string) string
	Print    func(string) string
	Error    func(error) string
}

var DefaultTheme = &Theme{
	Describe: NoopDecorator,
	Print:    NoopDecorator,
	Error:    NoopErrDecorator,
}

var (
	ErrThemeDescribeMissing = fmt.Errorf("theme missing describe decorator")
	ErrThemePrintMissing    = fmt.Errorf("theme missing print decorator")
	ErrThemeErrorMissing    = fmt.Errorf("theme missing error decorator")
)

func NoopDecorator(s string) string {
	return s
}

func NoopErrDecorator(err error) string {
	return err.Error()
}

func ValidateTheme(theme *Theme) error {
	if theme.Describe == nil {
		return fmt.Errorf("theme validation: %w", ErrThemeDescribeMissing)
	}

	if theme.Print == nil {
		return fmt.Errorf("theme validation: %w", ErrThemePrintMissing)
	}

	if theme.Error == nil {
		return fmt.Errorf("theme validation: %w", ErrThemeErrorMissing)
	}

	return nil
}
