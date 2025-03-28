package indexer

import (
	"log"
	"path/filepath"
	"search-indexer/running"
	"search-indexer/server/conf"
	"search-indexer/server/core/document"
	"search-indexer/server/core/storage"
	"search-indexer/utils"
	fsutils "search-indexer/utils/fs"
	gitutils "search-indexer/utils/git"
	"time"
)

type GitIgnoreFilter struct {
	ignore *gitutils.GitIgnore
}

func (f *GitIgnoreFilter) Match(path string, isDir bool) bool {
	return !f.ignore.IsIgnored(path, isDir)
}

func demo() {
	conf := conf.Get()
	baseDir := conf.ForTest.Path
	if baseDir == "" {
		log.Println("ForTest.Path is not set")
		return
	}

	log.Println("Indexing:", baseDir)

	var filter fsutils.ListFileFilter
	if conf.Filters.Exclude.UseGitIgnore {
		log.Println("Using gitignore filter")
		filter = &GitIgnoreFilter{
			ignore: gitutils.NewGitIgnore(baseDir),
		}
	} else {
		log.Println("Using customized filter")
		filter = utils.NewSimpleFilterExclude(conf.Filters.Exclude.Customized, baseDir)
	}

	files, err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{
		Filter: utils.NewSimpleFilterExclude(conf.Filters.Exclude.Customized, baseDir),
	})

	if err != nil {
		log.Println("Error listing files:", err)
		files = []fsutils.FileInfo{}
	}

	log.Println(len(files), "files found.")

	filter = utils.NewSimpleFilter(conf.Filters.Include, baseDir)
	filteredFiles := []string{}
	for _, file := range files {
		if running.IsShuttingDown() {
			return
		}

		baseName := filepath.Base(file.Path)
		if filter.Match(baseName, false) {
			filteredFiles = append(filteredFiles, file.Path)
		}
	}

	log.Println(len(filteredFiles), "files matched.")

	succ := 0
	faied := 0
	last := time.Now()
	wordCount := 0
	docs := []*document.Document{}
	for n, file := range filteredFiles {
		if running.IsShuttingDown() {
			return
		}

		doc, err := document.Parse(file, baseDir)
		if err != nil {
			faied++
		} else {
			succ++
			wordCount += len(doc.Content.Words)
		}
		docs = append(docs, doc)
		if len(docs) > 100 {
			storage.Save(docs, "0")
			docs = []*document.Document{}
		}

		if time.Since(last) > 200*time.Millisecond || n == len(filteredFiles)-1 {
			last = time.Now()
			log.Printf("Parsing progress %d / %d, succ: %d, failed, %d, wordCount: %d", n+1, len(filteredFiles), succ, faied, wordCount)
		}
	}

	if len(docs) > 0 {
		storage.Save(docs, "0")
	}

	log.Println(len(filteredFiles), "parsed files, succ:", succ, "failed:", faied, "wordCount:", wordCount)

}
