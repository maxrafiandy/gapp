package database

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type QueryOptions struct {
	Page    int
	Limit   int
	Offset  int
	Sort    []SortField
	Search  *SearchQuery
	Filters map[string]string
	Select  []string
	Between map[string][2]string
}

type SortField struct {
	Field string
	Desc  bool
}

type SearchQuery struct {
	Field   string
	Keyword string
}

// Pagination result
type Pagination struct {
	Items any   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// AllowedFieldProvider interface
type AllowedFieldProvider interface {
	AllowedFields() map[string]string
	Model() *gorm.DB
}

func DefaultAllowedFields(model any) map[string]string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// ambil nama dari json tag kalau ada, fallback ke nama field
		jsonName := f.Tag.Get("json")
		if jsonName == "" || jsonName == "-" {
			jsonName = strings.ToLower(f.Name)
		}

		// ambil kolom dari tag gorm
		gormTag := f.Tag.Get("gorm")
		var columnName string
		if gormTag != "" {
			for _, part := range strings.Split(gormTag, ";") {
				if strings.HasPrefix(part, "column:") {
					columnName = strings.TrimPrefix(part, "column:")
					break
				}
			}
		}

		if columnName != "" {
			fields[jsonName] = columnName
		}
	}
	return fields
}

// ParseOpts converts URL query into QueryOptions
func ParseOpts(values url.Values) QueryOptions {
	opts := QueryOptions{
		Page:    1,
		Limit:   10,
		Filters: map[string]string{},
		Between: map[string][2]string{},
	}

	if p := values.Get("page"); p != "" {
		if page, _ := strconv.Atoi(p); page > 0 {
			opts.Page = page
		}
	}
	if l := values.Get("limit"); l != "" {
		if limit, _ := strconv.Atoi(l); limit > 0 {
			opts.Limit = limit
		}
	}

	opts.Offset = (opts.Page - 1) * opts.Limit

	if sort := values.Get("sort"); sort != "" {
		fields := strings.Split(sort, ",")
		for _, f := range fields {
			desc := strings.HasPrefix(f, "-")
			field := strings.TrimPrefix(f, "-")
			opts.Sort = append(opts.Sort, SortField{Field: field, Desc: desc})
		}
	}

	if search := values.Get("search"); search != "" {
		parts := strings.SplitN(search, ",", 2)
		if len(parts) == 2 {
			opts.Search = &SearchQuery{Field: parts[0], Keyword: parts[1]}
		}
	}

	if sel := values.Get("select"); sel != "" {
		opts.Select = strings.Split(sel, ",")
	}

	// Filters & Between
	for key, val := range values {
		if key == "page" || key == "limit" || key == "sort" || key == "search" || key == "select" {
			continue
		}

		if strings.HasSuffix(key, "[]") && len(val) == 2 {
			opts.Between[strings.TrimSuffix(key, "[]")] = [2]string{val[0], val[1]}
			continue
		}

		opts.Filters[key] = val[0]
	}

	return opts
}

// ApplyQuery runs full query with filters, search, sort, etc.
func PaginationResult(model AllowedFieldProvider, opts QueryOptions, out any) (*Pagination, error) {
	db := model.Model()
	allowed := model.AllowedFields()

	// Filter
	for key, val := range opts.Filters {
		if col, ok := allowed[key]; ok {
			db = db.Where(fmt.Sprintf("%s = ?", col), val)
		}
	}

	// Between
	for key, val := range opts.Between {
		if col, ok := allowed[key]; ok {
			if val[0] != "" && val[1] != "" {
				db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", col), val[0], val[1])
			}
		}
	}

	// Search
	if opts.Search != nil {
		if col, ok := allowed[opts.Search.Field]; ok {
			db = db.Where(fmt.Sprintf("%s LIKE ?", col), "%"+opts.Search.Keyword+"%")
		}
	}

	// Select
	if len(opts.Select) > 0 {
		columns := []string{}
		for _, field := range opts.Select {
			if col, ok := allowed[field]; ok {
				columns = append(columns, col)
			}
		}
		if len(columns) > 0 {
			db = db.Select(columns)
		}
	}

	// Sort
	for _, s := range opts.Sort {
		if col, ok := allowed[s.Field]; ok {
			if s.Desc {
				db = db.Order(col + " DESC")
			} else {
				db = db.Order(col + " ASC")
			}
		}
	}

	// Count
	var total int64
	countDB := db.Session(&gorm.Session{})
	if err := countDB.Count(&total).Error; err != nil {
		return nil, err
	}

	// Pagination
	db = db.Offset(opts.Offset).Limit(opts.Limit)

	// Data
	if err := db.Find(out).Error; err != nil {
		return nil, err
	}

	return &Pagination{
		Items: out,
		Total: total,
		Page:  opts.Page,
		Limit: opts.Limit,
	}, nil
}
