package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
	"github.com/tjm/puppet-patching-automation/version"
)

var formatAllSupported = []string{gin.MIMEHTML, gin.MIMEJSON, gin.MIMEYAML, gin.MIMEXML}

// var formatHTMLOnly = []string{gin.MIMEHTML}

// GetHome endpoint (GET)
func GetHome(c *gin.Context) {
	session := sessions.Default(c)
	name := session.Get("name")
	data := gin.H{
		"version": version.FormattedVersion(),
		"name":    name,
	}

	//c.HTML(http.StatusOK, "index.tmpl", data)
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "index.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, models.GetDefaultBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// GetPing endpoint (GET)
// basic healthcheck
func GetPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// convertSliceStringToUint will convert a slice of strings to a slice of uints
func convertSliceStringToUint(ss []string) (si []uint, err error) {
	si = make([]uint, 0, len(ss))
	for _, s := range ss {
		var i64 uint64
		i64, err = strconv.ParseUint(s, 0, 64)
		if err != nil {
			log.Error("Error converting strings to uint: ", err)
			break
		}
		si = append(si, uint(i64))
	}
	return
}

// validateID ensures we have a uint
func validateID(c *gin.Context, idParam string) (id uint, err error) {
	idString := c.Param(idParam)
	if idString == "latest" {
		err = errIDLatest
		return
	} else if idString == "new" {
		err = errIDNew
		return
	}
	id64, err := strconv.ParseUint(idString, 0, 64)
	id = uint(id64)
	if id == 0 {
		err = errInvalidID
	}
	if err != nil {
		log.Error("Error retrieving/parsing id parameter: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving/parsing id parameter: " + errInvalidID.Error()})
		c.Abort()
	}
	return
}

func getHTMLData(c *gin.Context, breadcrumbs models.BreadCrumbs, data ...gin.H) (htmlData gin.H) {
	breadcrumbs[len(breadcrumbs)-1].Active = true // Set last breadcrumb as "Active"
	data = append(data, gin.H{
		"session":     sessions.Default(c),
		"breadcrumbs": breadcrumbs,
		"title":       breadcrumbs[len(breadcrumbs)-1].Name, // Name of last breadcrumb is title
	})
	htmlData = mergeData(data...)
	return
}

// mergeData will merge a list of gin.H's into one
func mergeData(data ...gin.H) gin.H {
	// If no data is provided, return an empty gin.H
	if len(data) == 0 {
		return gin.H{}
	}

	// Merge a single data? return it!
	if len(data) == 1 {
		return data[0]
	}

	// Merge data[1] into data[0]
	for k, v := range data[1] {
		data[0][k] = v
	}

	// If this was the last pair, return
	if len(data) == 2 {
		return data[0]
	}

	// If there are still more to process, create a new slice, without data[1]
	// ... and call myself (yay recursive!?)
	return mergeData(append(data[:0], data[2:]...)...)
}
