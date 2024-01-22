package credentials

// IsTokenProviderHappy helps other packages check if a TokenProvider is set up properly.
// It's a simple check but defining it here helps with encapsulation. It can handle nil
// inputs.
func IsTokenProviderHappy(tp TokenProvider) bool {
	if tp == nil {
		return false
	} else {
		token, err := tp.Get()
		return err == nil && len(token) > 0
	}
}
