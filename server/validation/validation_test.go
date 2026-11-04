package validation

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

func assertInvalid(t *testing.T, name string, err error, message string) {
	t.Helper()
	status := trevrpc.StatusFromError(err)
	if err == nil || status.Code != trevrpc.CodeInvalidArgument || status.Message != message {
		t.Errorf("%s = %v, want InvalidArgument %q", name, err, message)
	}
}

func assertValid(t *testing.T, name string, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("%s = %v, want nil", name, err)
	}
}

func TestFirst(t *testing.T) {
	first, second := errors.New("first"), errors.New("second")
	if err := First(nil, first, second); !errors.Is(err, first) {
		t.Errorf("First() = %v, want %v", err, first)
	}
	assertValid(t, "First()", First(nil, nil))
}

func TestRequired(t *testing.T) {
	assertValid(t, "Required(set)", Required("name", true))
	assertInvalid(t, "Required(unset)", Required("name", false), "name is required")
}

func TestLength(t *testing.T) {
	assertValid(t, "Length(minimum)", Length("name", "abc", 3, 5))
	assertValid(t, "Length(multibyte)", Length("name", "ééééé", 3, 5))
	assertInvalid(t, "Length(short)", Length("name", "ab", 3, 5), "name must be at least 3 characters")
	assertInvalid(t, "Length(long)", Length("name", "abcdef", 3, 5), "name must be at most 5 characters")
}

func TestMaxBytes(t *testing.T) {
	assertValid(t, "MaxBytes(maximum)", MaxBytes("password", "éé", 4))
	assertInvalid(t, "MaxBytes(long)", MaxBytes("password", "ééé", 4), "password must be at most 4 bytes")
}

func TestPattern(t *testing.T) {
	pattern := regexp.MustCompile(`^[a-z]+$`)
	assertValid(t, "Pattern(match)", Pattern("name", "abc", pattern, "be lowercase"))
	assertInvalid(t, "Pattern(mismatch)", Pattern("name", "ABC", pattern, "be lowercase"), "name must be lowercase")
}

func TestBetween(t *testing.T) {
	assertValid(t, "Between(minimum)", Between("number", 1, 1, 10))
	assertValid(t, "Between(maximum)", Between("number", 10, 1, 10))
	assertInvalid(t, "Between(low)", Between("number", 0, 1, 10), "number must be between 1 and 10")
	assertInvalid(t, "Between(high)", Between("number", 11, 1, 10), "number must be between 1 and 10")
}

func TestLess(t *testing.T) {
	low, high := uint32(1), uint32(2)
	assertValid(t, "Less(ordered)", Less("minimum", &low, "maximum", &high))
	assertValid(t, "Less(unset)", Less("minimum", nil, "maximum", &low))
	assertInvalid(t, "Less(equal)", Less("minimum", &low, "maximum", &low), "minimum must be less than maximum")
	assertInvalid(t, "Less(reversed)", Less("minimum", &high, "maximum", &low), "minimum must be less than maximum")
}

func TestTimeRange(t *testing.T) {
	start := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	end := timestamppb.New(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	assertValid(t, "TimeRange(ordered)", TimeRange(start, end))
	assertValid(t, "TimeRange(unset)", TimeRange(nil, end))
	assertInvalid(t, "TimeRange(equal)", TimeRange(start, start), "end must be after start")
	assertInvalid(t, "TimeRange(reversed)", TimeRange(end, start), "end must be after start")
	assertInvalid(t, "TimeRange(invalid)", TimeRange(&timestamppb.Timestamp{Nanos: -1}, nil), "start is not a valid timestamp")
}
