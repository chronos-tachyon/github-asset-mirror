package indexfile

type Release struct {
	ID      int64   `json:"id,omitempty"`
	Tag     string  `json:"tag"`
	Name    string  `json:"name,omitempty"`
	Body    string  `json:"body,omitempty"`
	Version Version `json:"version"`
	Assets  []Asset `json:"assets,omitempty"`
}

func (r Release) CompareTo(other Release) CompareResult {
	cmp := r.Version.CompareTo(other.Version)
	if cmp == EQ {
		cmp = CompareString(r.Tag, other.Tag)
	}
	if cmp == EQ {
		cmp = CompareInt64(r.ID, other.ID)
	}
	return cmp
}

var (
	_ ComparableTo[Release] = Release{}
)
