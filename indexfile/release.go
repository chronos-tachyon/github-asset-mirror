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

func (r Release) FirstMatchingAsset(fn func(Asset) bool) (Asset, bool) {
	for _, a := range r.Assets {
		if fn(a) {
			return a, true
		}
	}
	return Asset{}, false
}

func (r Release) MatchingAssets(fn func(Asset) bool) []Asset {
	out := make([]Asset, 0, len(r.Assets))
	for _, a := range r.Assets {
		if fn(a) {
			out = append(out, a)
		}
	}
	return out
}

var (
	_ ComparableTo[Release] = Release{}
)
