package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"sitehub/web"
)

type viewMenu struct {
	Name, Code, Path, Permission string
	Sort                         int
	Menu                         *bool
	Hidden, Transparent          bool
}

func discoverMenus() ([]MenuSeed, error) {
	var seeds []MenuSeed
	for _, system := range []string{"admin", "blog", "product"} {
		views, err := fs.Sub(web.Views, system+"/src/views")
		if err != nil {
			return nil, err
		}
		children, err := viewChildren(views, system, ".", "")
		if err != nil {
			return nil, fmt.Errorf("%s menu catalog: %w", system, err)
		}
		seeds = append(seeds, children...)
	}
	return seeds, nil
}

// Read siblings before descending so each directory is inserted before its children.
func viewChildren(views fs.FS, system, directory, parent string) ([]MenuSeed, error) {
	entries, err := fs.ReadDir(views, directory)
	if err != nil {
		return nil, err
	}
	type child struct {
		directory string
		viewMenu
	}
	var children []child
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "components" {
			continue
		}
		dir := path.Join(directory, entry.Name())
		config := child{directory: dir, viewMenu: viewMenu{Name: entry.Name(), Sort: 1000}}
		data, err := fs.ReadFile(views, path.Join(dir, "menu.json"))
		if err == nil {
			err = json.Unmarshal(data, &config.viewMenu)
		}
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%s/menu.json: %w", dir, err)
		}
		if config.Menu == nil || *config.Menu {
			children = append(children, config)
		}
	}
	sort.SliceStable(children, func(i, j int) bool { return children[i].Sort < children[j].Sort })
	var result []MenuSeed
	for _, item := range children {
		dir := item.directory
		_, err := fs.Stat(views, path.Join(dir, "index.vue"))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		page := err == nil
		seed := MenuSeed{Code: item.Code, Parent: parent, System: system, Name: item.Name, Path: item.Path, Permission: item.Permission, Type: 1, Hidden: item.Hidden}
		if page {
			seed.Type = 2
			seed.Component = dir + "/index"
		}
		if seed.Code == "" {
			seed.Code = system + "-" + strings.ReplaceAll(dir, "/", "-")
			if system == "admin" && page {
				seed.Code = "page-" + strings.ReplaceAll(seed.Component, "/", "-")
			}
		}
		if seed.Path == "" {
			seed.Path = "/" + dir
			if system == "admin" {
				seed.Path = "/dashboard" + seed.Path
			}
		}
		nextParent := parent
		if !item.Transparent {
			nextParent = seed.Code
		}
		nested, err := viewChildren(views, system, dir, nextParent)
		if err != nil {
			return nil, err
		}
		if !item.Transparent && (page || len(nested) > 0) {
			result = append(result, seed)
		}
		result = append(result, nested...)
	}
	return result, nil
}
