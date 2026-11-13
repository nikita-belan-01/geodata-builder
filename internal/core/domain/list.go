package domain

import (
	"fmt"
	"sort"
)

type List struct {
	Name      string
	Roots     []string
	Optimized bool
	Prefix    string
}

func (l List) DirName() string {
	return "rus_" + l.Name
}

var lists = map[string]List{
	"full": {
		Name:   "full",
		Roots:  []string{"category-ru"},
		Prefix: "rus_full_part_",
	},
	"optimized": {
		Name:      "optimized",
		Roots:     []string{"category-ru"},
		Optimized: true,
		Prefix:    "rus_optimized_part_",
	},
	"banks": {
		Name:      "banks",
		Roots:     []string{"category-bank-ru"},
		Optimized: true,
		Prefix:    "rus_banks_part_",
	},
	"streaming": {
		Name:      "streaming",
		Roots:     []string{"category-media-ru", "category-entertainment-ru", "kinopoisk", "okko", "wink", "rutube"},
		Optimized: true,
		Prefix:    "rus_streaming_part_",
	},
}

func LookupList(name string) (List, error) {
	list, ok := lists[name]
	if !ok {
		return List{}, fmt.Errorf("%w: %q (available: %v)", ErrUnknownList, name, ListNames())
	}

	return list, nil
}

func Lists() []List {
	all := make([]List, 0, len(lists))
	for _, list := range lists {
		all = append(all, list)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })

	return all
}

func ListNames() []string {
	names := make([]string, 0, len(lists))
	for name := range lists {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}
