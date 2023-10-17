package models

import (
	"errors"
	"fmt"
	"strconv"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/tjm/puppet-patching-automation/functions"
)

// JenkinsJobParam defines JenkinsJobParamParams
type JenkinsJobParam struct {
	gorm.Model
	Name          string
	Type          string
	Description   string
	DefaultString string
	DefaultBool   bool
	TemplateValue string
	IsNotInJob    bool               // Finds old parameters
	template      *template.Template `gorm:"-"`
	JenkinsJobID  uint               `form:"-"`                           // Parent Jenkins Job ID
	JenkinsJob    *JenkinsJob        `json:"-" yaml:"-" xml:"-" form:"-"` // Parent Jenkins Job
}

// NewJenkinsJobParam returns a new JenkinsJobParam object
func NewJenkinsJobParam() (j *JenkinsJobParam) {
	j = new(JenkinsJobParam)
	// Set Defaults
	return
}

// Init : Create new PatchRun object
func (j *JenkinsJobParam) Init() error {
	err := j.buildTemplate()
	if err != nil {
		return err
	}
	return GetDB().Create(j).Error
}

// Save : Save PatchRun object
func (j *JenkinsJobParam) Save() error {
	err := j.buildTemplate()
	if err != nil {
		return err
	}
	return GetDB().Save(j).Error
}

// Delete : Delete PatchRun object
func (j *JenkinsJobParam) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// No Child Objects yet
	// 	// TODO: DO something about the template?
	// }
	return GetDB().Delete(j).Error
}

// GetTemplate : Returns Template for Param
func (j *JenkinsJobParam) GetTemplate() (tpl *template.Template, err error) {
	if j.template == nil {
		err = j.buildTemplate()
		if err != nil {
			return // logged in buildTemplate
		}
	}
	return j.template, nil
}

// SetDefaultValue will set the appropriate Default value type
func (j *JenkinsJobParam) SetDefaultValue(value interface{}) (err error) {
	switch j.Type {
	case "":
		return errors.New("type must be set before trying to set a default value")
	case "BooleanParameterDefinition":
		j.DefaultBool = value.(bool)
	case "StringParameterDefinition", "ChoiceParameterDefinition", "TextParameterDefinition":
		j.DefaultString = value.(string)
	default:
		return errors.New("Unable to handle parameter type: " + j.Type)
	}
	return
}

// GetDefaultValue will get the appropriate Default value based on the type
func (j *JenkinsJobParam) GetDefaultValue() (value string, err error) {
	switch j.Type {
	case "":
		err = errors.New("type must be set before trying to get a default value")
	case "BooleanParameterDefinition":
		value = strconv.FormatBool(j.DefaultBool)
	case "StringParameterDefinition", "ChoiceParameterDefinition", "TextParameterDefinition":
		value = j.DefaultString
	default:
		err = errors.New("Unable to handle parameter type: " + j.Type)
	}
	return
}

// buildTemplate : Builds Template
func (j *JenkinsJobParam) buildTemplate() (err error) {
	templateName := fmt.Sprintf("JobParamID-%v", j.ID) // I am not even sure where this us used.
	tpl, err := template.New(templateName).Funcs(template.FuncMap{
		"FormatAsDateTimeLocal": functions.FormatAsDateTimeLocal,
		"FormatAsISO8601":       functions.FormatAsISO8601,
	}).Parse(j.TemplateValue)
	if err != nil {
		log.WithFields(log.Fields{
			"ID":        j.ID,
			"paramName": j.Name,
		}).Error("Error Parsing Template: ", err)
		return
	}
	j.template = tpl
	return
}

// GetJenkinsJobParamByID returns patch run object by ID
func GetJenkinsJobParamByID(id uint) (j *JenkinsJobParam, err error) {
	j = new(JenkinsJobParam)
	err = GetDB().First(j, id).Error
	return
}

// GetJenkinsJobParams returns a list of all JenkinsJobParams
func GetJenkinsJobParams() (servers []*JenkinsJobParam) {
	servers = make([]*JenkinsJobParam, 0)
	GetDB().Order("name").Find(&servers)
	return
}
