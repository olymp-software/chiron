// Package source handles course sources: normalizing them, deriving storage
// keys, copying local folders and fetching course repositories.
package source

import (
	"slices"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CloneTags clones the repository at url into dir without checking out a
// working tree and returns its tag names, sorted.
func CloneTags(url, dir string) ([]string, error) {
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:        url,
		NoCheckout: true,
		Tags:       git.AllTags,
	})
	if err != nil {
		return nil, err
	}
	iter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	var tags []string
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		tags = append(tags, ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(tags)
	return tags, nil
}
