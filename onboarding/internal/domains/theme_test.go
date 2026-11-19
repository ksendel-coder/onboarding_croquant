package domains

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeThemePatch(t *testing.T) {
	valid := "#a1b2c3"
	result, err := NormalizeThemePatch(ThemePatch{PrimaryColor: &valid})
	require.NoError(t, err)
	require.Equal(t, "#A1B2C3", *result.PrimaryColor)
	require.Nil(t, result.BackgroundColor)
	require.Nil(t, result.TextColor)
	require.Equal(t, "#a1b2c3", valid)
	for _, value := range []string{"", "red", "#ABC", "#12345678", "#GGGGGG", " #FFFFFF", "url(x)"} {
		t.Run(value, func(t *testing.T) {
			for _, patch := range []ThemePatch{{PrimaryColor: &value}, {BackgroundColor: &value}, {TextColor: &value}} {
				_, err := NormalizeThemePatch(patch)
				require.ErrorIs(t, err, ErrInvalidThemeColor)
			}
		})
	}
}
