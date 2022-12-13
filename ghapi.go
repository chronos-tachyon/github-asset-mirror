package main

import (
	"github.com/google/go-github/v48/github"
)

type CallFunc[T any] func(*github.ListOptions) ([]*T, *github.Response)

type ProcessFunc[T any] func(*T)

func Iterate[T any](pageSize int, callFn CallFunc[T], processFn ProcessFunc[T]) {
	var options github.ListOptions
	options.Page = 0
	options.PerPage = pageSize
	for {
		list, resp := callFn(&options)
		for _, item := range list {
			processFn(item)
		}
		if resp.NextPage == 0 {
			return
		}
		options.Page = resp.NextPage
	}
}
