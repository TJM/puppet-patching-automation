package models

// BreadCrumb defines a BreadCrumb, use for navigation
type BreadCrumb struct {
	Name   string `binding:"required"`
	URL    string `binding:"required,url"`
	Active bool
}

// BreadCrumbs is a list of BreadCrumb objects
type BreadCrumbs []*BreadCrumb

// createBreadCrumb will return the BreadCrumb object from the components
func createBreadCrumb(name, url string) *BreadCrumb {
	return &BreadCrumb{
		Name: name,
		URL:  url,
	}
}

// GetDefaultBreadCrumbs will return the default breadcrumbs (top level)
func GetDefaultBreadCrumbs() BreadCrumbs {
	return BreadCrumbs{createBreadCrumb("Home", "/")}
}
