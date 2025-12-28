// Package validator provides a pluggable validator registry for field validation
package validator

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"sync"
)

// ValidatorFunc is the signature for validator functions
// Takes a value and optional configuration, returns an error if validation fails
type ValidatorFunc func(value interface{}, config map[string]interface{}) error

// Registry holds registered validators
type Registry struct {
	validators map[string]ValidatorFunc
	mu         sync.RWMutex
}

var (
	defaultRegistry *Registry
	once            sync.Once
)

// GetRegistry returns the singleton validator registry
func GetRegistry() *Registry {
	once.Do(func() {
		defaultRegistry = &Registry{
			validators: make(map[string]ValidatorFunc),
		}
		defaultRegistry.registerBuiltins()
	})
	return defaultRegistry
}

// Register adds a validator to the registry
func (r *Registry) Register(name string, fn ValidatorFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.validators[name] = fn
}

// Get returns a validator by name
func (r *Registry) Get(name string) (ValidatorFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.validators[name]
	return fn, ok
}

// Validate runs a named validator
func (r *Registry) Validate(name string, value interface{}, config map[string]interface{}) error {
	fn, ok := r.Get(name)
	if !ok {
		return fmt.Errorf("validator '%s' not found", name)
	}
	return fn(value, config)
}

// List returns all registered validator names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.validators))
	for name := range r.validators {
		names = append(names, name)
	}
	return names
}

// registerBuiltins registers all built-in validators
func (r *Registry) registerBuiltins() {
	// Email validator
	r.Register("email", func(value interface{}, config map[string]interface{}) error {
		str, ok := value.(string)
		if !ok || str == "" {
			return nil // Empty values handled by required check
		}
		_, err := mail.ParseAddress(str)
		if err != nil {
			return fmt.Errorf("invalid email format")
		}
		return nil
	})

	// URL validator
	r.Register("url", func(value interface{}, config map[string]interface{}) error {
		str, ok := value.(string)
		if !ok || str == "" {
			return nil
		}
		// Simple URL validation
		if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
			return fmt.Errorf("URL must start with http:// or https://")
		}
		return nil
	})

	// Phone validator (basic)
	r.Register("phone", func(value interface{}, config map[string]interface{}) error {
		str, ok := value.(string)
		if !ok || str == "" {
			return nil
		}
		// Allow digits, spaces, dashes, parentheses, and plus
		cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(str, "")
		if len(cleaned) < 7 || len(cleaned) > 15 {
			return fmt.Errorf("phone number must have 7-15 digits")
		}
		return nil
	})

	// Regex validator
	r.Register("regex", func(value interface{}, config map[string]interface{}) error {
		str, ok := value.(string)
		if !ok || str == "" {
			return nil
		}
		pattern, _ := config["pattern"].(string)
		if pattern == "" {
			return nil
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %v", err)
		}
		if !re.MatchString(str) {
			if msg, ok := config["message"].(string); ok && msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("value does not match required pattern")
		}
		return nil
	})

	// Length validator
	r.Register("length", func(value interface{}, config map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return nil
		}
		length := len(str)
		if min, ok := config["min"].(float64); ok && length < int(min) {
			return fmt.Errorf("must be at least %d characters", int(min))
		}
		if max, ok := config["max"].(float64); ok && length > int(max) {
			return fmt.Errorf("must be at most %d characters", int(max))
		}
		return nil
	})

	// Numeric range validator
	r.Register("range", func(value interface{}, config map[string]interface{}) error {
		var num float64
		switch v := value.(type) {
		case float64:
			num = v
		case int:
			num = float64(v)
		case int64:
			num = float64(v)
		default:
			return nil
		}
		if min, ok := config["min"].(float64); ok && num < min {
			return fmt.Errorf("must be at least %.2f", min)
		}
		if max, ok := config["max"].(float64); ok && num > max {
			return fmt.Errorf("must be at most %.2f", max)
		}
		return nil
	})

	// Alphanumeric validator
	r.Register("alphanumeric", func(value interface{}, config map[string]interface{}) error {
		str, ok := value.(string)
		if !ok || str == "" {
			return nil
		}
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, str)
		if !matched {
			return fmt.Errorf("must contain only letters and numbers")
		}
		return nil
	})

	// Credit card validator (Luhn algorithm)
	r.Register("creditcard", func(value interface{}, config map[string]interface{}) error {
		str, ok := value.(string)
		if !ok || str == "" {
			return nil
		}
		// Remove non-digits
		digits := regexp.MustCompile(`[^\d]`).ReplaceAllString(str, "")
		if len(digits) < 13 || len(digits) > 19 {
			return fmt.Errorf("invalid credit card number length")
		}
		// Luhn check
		if !luhnCheck(digits) {
			return fmt.Errorf("invalid credit card number")
		}
		return nil
	})
}

// luhnCheck performs Luhn algorithm validation
func luhnCheck(number string) bool {
	sum := 0
	parity := len(number) % 2
	for i, r := range number {
		digit := int(r - '0')
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}

// Package-level convenience functions

// Register adds a validator to the default registry
func Register(name string, fn ValidatorFunc) {
	GetRegistry().Register(name, fn)
}

// Validate runs a named validator using the default registry
func Validate(name string, value interface{}, config map[string]interface{}) error {
	return GetRegistry().Validate(name, value, config)
}
