package v1

import (
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	numberv1 "trev.zip/template/stack/server/gen/number/v1"
	"trev.zip/template/stack/server/validation"
)

const (
	maxNameLength = 50
	maxNumber     = 1_000_000
	maxBuckets    = 100
	maxNames      = 100
)

func validateAddRequest(req *numberv1.AddRequest) error {
	return validation.First(
		validation.Required("name", req.HasName()),
		validation.Length("name", req.GetName(), 1, maxNameLength),
		validation.Required("number", req.HasNumber()),
		validation.Between("number", req.GetNumber(), 1, maxNumber),
	)
}

func validateListRequest(req *numberv1.ListRequest) error {
	var minimum, maximum *uint32
	if req.HasMinimum() {
		minimum = new(req.GetMinimum())
	}
	if req.HasMaximum() {
		maximum = new(req.GetMaximum())
	}
	return validation.First(
		validation.Length("name", req.GetName(), 0, maxNameLength),
		validation.Between("minimum", req.GetMinimum(), 0, maxNumber),
		validation.Between("maximum", req.GetMaximum(), 0, maxNumber),
		validation.Less("minimum", minimum, "maximum", maximum),
		validation.TimeRange(req.GetStart(), req.GetEnd()),
	)
}

func validateTimeSeriesRequest(req *numberv1.TimeSeriesRequest) error {
	if err := validation.Required("interval", req.HasInterval()); err != nil {
		return err
	}
	interval := req.GetInterval()
	if _, defined := numberv1.TimeInterval_name[int32(interval)]; !defined ||
		interval == numberv1.TimeInterval_TIME_INTERVAL_UNSPECIFIED {
		return trevrpc.InvalidArgument("interval must be hour, day, week, or month")
	}
	return validation.TimeRange(req.GetStart(), req.GetEnd())
}

func validateDistributionRequest(req *numberv1.DistributionRequest) error {
	return validation.First(
		validation.Required("bucket_count", req.HasBucketCount()),
		validation.Between("bucket_count", req.GetBucketCount(), 1, maxBuckets),
		validation.TimeRange(req.GetStart(), req.GetEnd()),
	)
}

func validateTopNamesRequest(req *numberv1.TopNamesRequest) error {
	return validation.First(
		validation.Required("limit", req.HasLimit()),
		validation.Between("limit", req.GetLimit(), 1, maxNames),
		validation.TimeRange(req.GetStart(), req.GetEnd()),
	)
}
