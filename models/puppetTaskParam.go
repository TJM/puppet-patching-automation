package models

import (
	"fmt"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/tjm/puppet-patching-automation/functions"
)

// PuppetTaskParam defines PuppetTaskParamParams
type PuppetTaskParam struct {
	gorm.Model
	Name          string
	Type          string
	Description   string
	TemplateValue string
	IsNotInTask   bool               // Finds old parameters
	template      *template.Template `gorm:"-" binding:"-" form:"-"`
	PuppetTaskID  uint               `form:"-"`                           // Parent PuppetTask ID
	PuppetTask    *PuppetTask        `json:"-" yaml:"-" xml:"-" form:"-"` // Parent PuppetTask
}

// NewPuppetTaskParam returns a new PuppetTaskParam object
func NewPuppetTaskParam() (p *PuppetTaskParam) {
	p = new(PuppetTaskParam)
	// Set Defaults
	return
}

// Init : Create new PatchRun object
func (p *PuppetTaskParam) Init() error {
	err := p.buildTemplate()
	if err != nil {
		return err
	}
	return GetDB().Create(p).Error
}

// Save : Save PatchRun object
func (p *PuppetTaskParam) Save() error {
	err := p.buildTemplate()
	if err != nil {
		return err
	}
	return GetDB().Save(p).Error
}

// Delete : Delete PatchRun object
func (p *PuppetTaskParam) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// No Child Objects yet
	// 	// TODO: DO something about the template?
	// }
	return GetDB().Delete(p).Error
}

// GetTemplate : Returns Template for Param
func (p *PuppetTaskParam) GetTemplate() (tpl *template.Template, err error) {
	if p.template == nil {
		err = p.buildTemplate()
		if err != nil {
			return // logged in buildTemplate
		}
	}
	return p.template, nil
}

// GetDefaultValue will get the appropriate Default value based on the type
// TODO: See if there is a way to get the default value from Puppet API, for now returns ""
func (p *PuppetTaskParam) GetDefaultValue() (value string, err error) {
	// switch p.Type {
	// case "":
	// 	err = errors.New("type must be set before trying to get a default value")
	// case "BooleanParameterDefinition":
	// 	value = strconv.FormatBool(p.DefaultBool)
	// case "StringParameterDefinition", "ChoiceParameterDefinition", "TextParameterDefinition":
	// 	value = p.DefaultString
	// default:
	// 	err = errors.New("Unable to handle parameter type: " + p.Type)
	// }
	return
}

// buildTemplate : Builds Template
func (p *PuppetTaskParam) buildTemplate() (err error) {
	templateName := fmt.Sprintf("PuppetTaskParamID-%v", p.ID) // I am not even sure where this us used.
	tpl, err := template.New(templateName).Funcs(template.FuncMap{
		"FormatAsDateTimeLocal": functions.FormatAsDateTimeLocal,
		"FormatAsISO8601":       functions.FormatAsISO8601,
	}).Parse(p.TemplateValue)
	if err != nil {
		log.WithFields(log.Fields{
			"ID":        p.ID,
			"paramName": p.Name,
		}).Error("Error Parsing Template: ", err)
		return
	}
	p.template = tpl
	return
}

// GetPuppetTaskParamByID returns patch run object by ID
func GetPuppetTaskParamByID(id uint) (p *PuppetTaskParam, err error) {
	p = new(PuppetTaskParam)
	err = GetDB().First(p, id).Error
	return
}

// GetPuppetTaskParams returns a list of all PuppetTaskParams
func GetPuppetTaskParams() (params []*PuppetTaskParam) {
	params = make([]*PuppetTaskParam, 0)
	GetDB().Order("name").Find(&params)
	return
}
