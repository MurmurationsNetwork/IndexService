package model_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MurmurationsNetwork/MurmurationsServices/services/index/config"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/index/internal/model"
)

func TestConvertGeolocation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  map[string]interface{}
		expectErr bool
	}{
		{
			name: "valid geolocation string",
			input: `{
				"geolocation": "40.748817,-73.985428"
			}`,
			expected: map[string]interface{}{
				"geolocation": "40.748817,-73.985428",
				"latitude":    40.748817,
				"longitude":   -73.985428,
			},
			expectErr: false,
		},
		{
			name:      "invalid geolocation string",
			input:     `{"geolocation": "invalid_string"}`,
			expected:  map[string]interface{}{"geolocation": "invalid_string"},
			expectErr: true,
		},
		{
			name:      "geolocation not present",
			input:     `{}`,
			expected:  map[string]interface{}{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := model.NewTestProfile(tt.input)
			err := profile.ConvertGeolocation()
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, profile.GetJSON())
			}
		})
	}
}

func TestRepackageGeolocation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name: "both latitude and longitude are present",
			input: `{
				"latitude":  40.748817,
				"longitude": -73.985428
			}`,
			expected: map[string]interface{}{
				"latitude":  40.748817,
				"longitude": -73.985428,
				"geolocation": map[string]interface{}{
					"lat": 40.748817,
					"lon": -73.985428,
				},
			},
		},
		{
			name: "only latitude is present",
			input: `{
				"latitude": 40.748817
			}`,
			expected: map[string]interface{}{
				"latitude": 40.748817,
				"geolocation": map[string]interface{}{
					"lat": 40.748817,
					"lon": 0,
				},
			},
		},
		{
			name: "only longitude is present",
			input: `{
				"longitude": -73.985428
			}`,
			expected: map[string]interface{}{
				"longitude": -73.985428,
				"geolocation": map[string]interface{}{
					"lat": 0,
					"lon": -73.985428,
				},
			},
		},
		{
			name:     "neither latitude nor longitude are present",
			input:    `{}`,
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := model.NewTestProfile(tt.input)
			profile.RepackageGeolocation()
			require.Equal(t, tt.expected, profile.GetJSON())
		})
	}
}

func TestFilterTags(t *testing.T) {
	tests := []struct {
		name             string
		profileStr       string
		tagsArraySize    int
		tagsStringLength int
		expected         map[string]interface{}
		expectErr        bool
	}{
		{
			name:             "valid tags filtering",
			profileStr:       `{"tags": ["tag1", "tag2", "tag3"]}`,
			tagsArraySize:    3,
			tagsStringLength: 10,
			expected: map[string]interface{}{
				"tags": []string{"tag1", "tag2", "tag3"},
			},
			expectErr: false,
		},
		{
			name:             "invalid tags filtering, too many tags",
			profileStr:       `{"tags": ["tag1", "tag2", "tag3"]}`,
			tagsArraySize:    2,
			tagsStringLength: 10,
			expected: map[string]interface{}{
				"tags": []string{"tag1", "tag2"},
			},
			expectErr: false,
		},
		{
			name:             "valid tags filtering, truncate long tag",
			profileStr:       `{"tags": ["verylongtagwithmorethan10chars", "tag2"]}`,
			tagsArraySize:    2,
			tagsStringLength: 10,
			expected: map[string]interface{}{
				"tags": []string{"verylongta", "tag2"},
			},
			expectErr: false,
		},
		{
			name:             "invalid profileStr",
			profileStr:       `{tags: ["tag1", "tag2"]}`, // incorrect JSON syntax
			tagsArraySize:    3,
			tagsStringLength: 10,
			expected:         map[string]interface{}{},
			expectErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Values.Server.TagsArraySize = strconv.Itoa(tt.tagsArraySize)
			config.Values.Server.TagsStringLength = strconv.Itoa(
				tt.tagsStringLength,
			)

			profile := model.NewTestProfile(tt.profileStr)
			err := profile.FilterTags()
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, profile.GetJSON())
			}
		})
	}
}

func TestValidatePrimaryURL(t *testing.T) {
	tests := []struct {
		name       string
		primaryURL string
		expected   map[string]interface{}
		expectErr  bool
	}{
		{
			name:       "valid URL",
			primaryURL: "https://example.com/path?query=param",
			expected: map[string]interface{}{
				"primary_url": "example.com/path?query=param",
			},
			expectErr: false,
		},
		{
			name:       "valid URL with www",
			primaryURL: "https://www.example.com/path?query=param",
			expected: map[string]interface{}{
				"primary_url": "example.com/path?query=param",
			},
			expectErr: false,
		},
		{
			name:       "valid URL without scheme",
			primaryURL: "example.com/path?query=param",
			expected: map[string]interface{}{
				"primary_url": "example.com/path?query=param",
			},
			expectErr: false,
		},
		{
			name:       "invalid URL",
			primaryURL: "://example.com",
			expected: map[string]interface{}{
				"primary_url": "",
			},
			expectErr: false,
		},
		{
			name:       "no primary_url",
			primaryURL: "",
			expected: map[string]interface{}{
				"primary_url": "",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := model.NewTestProfile(
				`{"primary_url":"` + tt.primaryURL + `"}`,
			)

			err := profile.ValidatePrimaryURL()

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, profile.GetJSON())
			}
		})
	}
}

func TestSetDefaultStatus(t *testing.T) {
	profile := model.NewTestProfile("{}")
	profile.SetDefaultStatus()
	require.Equal(
		t,
		map[string]interface{}{"status": "posted"},
		profile.GetJSON(),
	)
}
