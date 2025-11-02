package api

type FileToSend struct {
	Name     string
	URL      string
	Username string
}

type MediaSearcher interface {
	Search(tags []string) (files []FileToSend, err error)
	FormatAndModifySearch(tags []string, authorID int64) (searchTerm string, err error)
}
