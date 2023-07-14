package model

// TestProfile is a wrapper around the Profile type that exposes its unexported
// methods for testing.
type TestProfile struct {
	*Profile
}

// NewTestProfile creates a new TestProfile instance.
func NewTestProfile(profileStr string) *TestProfile {
	return &TestProfile{
		NewProfile(profileStr),
	}
}

// ConvertGeolocation is a wrapper around the unexported convertGeolocation
// method in the Profile type.
func (s *TestProfile) ConvertGeolocation() error {
	return s.convertGeolocation()
}

// RepackageGeolocation is a wrapper around the unexported repackageGeolocation
// method in the Profile type.
func (s *TestProfile) RepackageGeolocation() {
	s.repackageGeolocation()
}

// NormalizeCountryCode is a wrapper around the unexported normalizeCountryCode
// method in the Profile type.
func (s *TestProfile) NormalizeCountryCode() error {
	return s.normalizeCountryCode()
}

// FilterTags is a wrapper around the unexported filterTags method in the Profile
// type.
func (s *TestProfile) FilterTags() error {
	return s.filterTags()
}

// ValidatePrimaryURL is a wrapper around the unexported validatePrimaryURL method
// in the Profile type.
func (s *TestProfile) ValidatePrimaryURL() error {
	return s.validatePrimaryURL()
}

// SetDefaultStatus is a wrapper around the unexported setDefaultStatus method in
// the Profile type.
func (s *TestProfile) SetDefaultStatus() {
	s.setDefaultStatus()
}
