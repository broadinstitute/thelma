package terra

type ReleaseFilter interface {
	Matches(Release) bool
	And(ReleaseFilter) ReleaseFilter
	Or(ReleaseFilter) ReleaseFilter
}

type filter struct {
	matcher func(Release) bool
}

func (f *filter) Matches(release Release) bool {
	return f.matcher(release)
}

func HasName(releaseName string) ReleaseFilter {
	return &filter{
		matcher: func(r Release) bool {
			return r.Name() == releaseName
		},
	}
}

func HasDestination(destinationName string) ReleaseFilter {
	return &filter{
		matcher: func(r Release) bool {
			return r.Destination().Name() == destinationName
		},
	}
}

func AnyRelease() ReleaseFilter {
	return &filter{
		matcher: func(_ Release) bool {
			return true
		},
	}
}

func (f *filter) And(other ReleaseFilter) ReleaseFilter {
	return &filter{
		matcher: func(release Release) bool {
			return f.Matches(release) && other.Matches(release)
		},
	}
}

func (f *filter) Or(other ReleaseFilter) ReleaseFilter {
	return &filter{
		matcher: func(release Release) bool {
			return f.Matches(release) || other.Matches(release)
		},
	}
}
