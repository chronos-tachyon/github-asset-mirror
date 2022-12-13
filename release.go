package main

type Release struct {
	ID      int64   `json:"id,omitempty"`
	Tag     string  `json:"tag"`
	Name    string  `json:"name,omitempty"`
	Body    string  `json:"body,omitempty"`
	Version Version `json:"version"`
	Assets  []Asset `json:"assets,omitempty"`
}
