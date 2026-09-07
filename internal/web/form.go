package web

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"
)

const (
	formMediaType   = "application/x-www-form-urlencoded"
	maxFormBodySize = 64 << 10 // 64 KiB
	maxTextLength   = 4096
	minOutputWidth  = 20
	maxOutputWidth  = 300
	defaultWidth    = 80
)

// formRequestError identifies request parsing and validation failures.
// Keeping the status alongside a safe client message lets every form handler
// apply the same HTTP contract without matching against error text.
type formRequestError struct {
	status  int
	message string
	err     error
}

func (e *formRequestError) Error() string {
	if e.err == nil {
		return e.message
	}
	return fmt.Sprintf("%s: %v", e.message, e.err)
}

func (e *formRequestError) Unwrap() error {
	return e.err
}

func badFormRequest(message string) error {
	return &formRequestError{
		status:  http.StatusBadRequest,
		message: message,
	}
}

// parsePostForm applies the request-level security boundaries shared by the
// generation and download endpoints. Field cardinality and values are
// validated separately after this function succeeds.
func parsePostForm(w http.ResponseWriter, r *http.Request) (url.Values, error) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != formMediaType {
		return nil, &formRequestError{
			status:  http.StatusUnsupportedMediaType,
			message: "Form submissions must use URL-encoded data.",
			err:     err,
		}
	}

	// The limit must wrap the body before ParseForm reads it. MaxBytesReader
	// reads at most one byte beyond the limit so an oversized body can be
	// distinguished from other malformed form data.
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodySize)
	if err := r.ParseForm(); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return nil, &formRequestError{
				status:  http.StatusRequestEntityTooLarge,
				message: "The submitted form is too large.",
				err:     err,
			}
		}
		return nil, &formRequestError{
			status:  http.StatusBadRequest,
			message: "The submitted form is malformed.",
			err:     err,
		}
	}

	// PostForm contains body values only. Returning Form here would merge in
	// query parameters and allow a URL value to conflict with submitted data.
	return r.PostForm, nil
}

// parseGenerationForm composes the request and field-validation stages used by
// both POST endpoints. State is returned even on field errors so handlers can
// redisplay every unambiguous submitted value.
func parseGenerationForm(w http.ResponseWriter, r *http.Request) (FormState, GenerationInput, error) {
	values, err := parsePostForm(w, r)
	if err != nil {
		return defaultFormState(), GenerationInput{}, err
	}

	state, err := decodeForm(values)
	if err != nil {
		return state, GenerationInput{}, err
	}

	input, err := validateForm(state)
	if err != nil {
		return state, GenerationInput{}, err
	}
	return state, input, nil
}

// decodeForm checks the submitted key set and field cardinality before copying
// raw values into FormState. The returned state is still untrusted and must be
// passed to validateForm before generation.
func decodeForm(values url.Values) (FormState, error) {
	state := FormState{
		Text:      singleFormValue(values, "text"),
		Banner:    singleFormValue(values, "banner"),
		Color:     singleFormValue(values, "color"),
		Substring: singleFormValue(values, "substring"),
		Alignment: singleFormValue(values, "align"),
		Width:     singleFormValue(values, "width"),
	}
	if _, submitted := values["width"]; !submitted {
		state.Width = strconv.Itoa(defaultWidth)
	}

	for key := range values {
		if !isAcceptedFormKey(key) {
			return state, badFormRequest("The submitted form contains an unexpected field.")
		}
	}

	keys := [...]string{"text", "banner", "use_color", "color", "substring", "align", "width"}
	for _, key := range keys {
		if submitted, ok := values[key]; ok && len(submitted) != 1 {
			return state, badFormRequest("Each form field must be submitted only once.")
		}
	}

	if submitted, ok := values["use_color"]; ok {
		if len(submitted) != 1 || submitted[0] != "on" {
			return state, badFormRequest("The color option is invalid.")
		}
		state.UseColor = true
	}

	return state, nil
}

// validateForm converts raw form state into the bounded values accepted by the
// shared generation workflow. It deliberately does not trim or normalize user
// text; those transformations would change established renderer semantics.
func validateForm(state FormState) (GenerationInput, error) {
	if state.Text == "" {
		return GenerationInput{}, badFormRequest("Text is required.")
	}
	if utf8.RuneCountInString(state.Text) > maxTextLength {
		return GenerationInput{}, badFormRequest("Text must be 4,096 characters or fewer.")
	}

	if !isSupportedBanner(state.Banner) {
		return GenerationInput{}, badFormRequest("Select a valid banner.")
	}
	if utf8.RuneCountInString(state.Substring) > maxTextLength {
		return GenerationInput{}, badFormRequest("Substring must be 4,096 characters or fewer.")
	}
	if !isSupportedAlignment(state.Alignment) {
		return GenerationInput{}, badFormRequest("Select a valid alignment.")
	}

	width, err := strconv.Atoi(state.Width)
	if err != nil || width < minOutputWidth || width > maxOutputWidth {
		return GenerationInput{}, badFormRequest("Width must be a whole number from 20 through 300.")
	}

	color := SelectedColor{Enabled: state.UseColor}
	if state.UseColor {
		if !isWebHexColor(state.Color) {
			return GenerationInput{}, badFormRequest("Select a valid color in #RRGGBB format.")
		}
		// C2 turns the validated hex string into bounded RGB integers, so the
		// template can build fixed CSS without trusting raw request text.
		convertedColor, err := selectedColorFromHex(state.Color)
		if err != nil {
			return GenerationInput{}, badFormRequest("Select a valid color in #RRGGBB format.")
		}
		color = convertedColor
	}

	return GenerationInput{
		Text:      state.Text,
		Banner:    state.Banner,
		Color:     color,
		Substring: state.Substring,
		Alignment: state.Alignment,
		Width:     width,
	}, nil
}

func singleFormValue(values url.Values, key string) string {
	if len(values[key]) != 1 {
		return ""
	}
	return values[key][0]
}

func isAcceptedFormKey(key string) bool {
	switch key {
	case "text", "banner", "use_color", "color", "substring", "align", "width":
		return true
	default:
		return false
	}
}

func isSupportedBanner(value string) bool {
	switch value {
	case BannerStandard, BannerShadow, BannerThinkertoy:
		return true
	default:
		return false
	}
}

func isSupportedAlignment(value string) bool {
	switch value {
	case "left", "center", "right", "justify":
		return true
	default:
		return false
	}
}

func isWebHexColor(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}
	for _, digit := range value[1:] {
		if !isHexDigit(digit) {
			return false
		}
	}
	return true
}

func isHexDigit(value rune) bool {
	return value >= '0' && value <= '9' ||
		value >= 'a' && value <= 'f' ||
		value >= 'A' && value <= 'F'
}
