package models

import (
	"fmt"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/tjm/puppet-patching-automation/functions"
)

// PuppetPlanParam defines PuppetPlanParamParams
type PuppetPlanParam struct {
	gorm.Model
	Name          string
	Type          string
	Description   string
	TemplateValue string
	IsNotInPlan   bool               // Finds old parameters
	template      *template.Template `gorm:"-" binding:"-" form:"-"`
	PuppetPlanID  uint               `form:"-"`                           // Parent PuppetPlan ID
	PuppetPlan    *PuppetPlan        `json:"-" yaml:"-" xml:"-" form:"-"` // Parent PuppetPlan
}

// NewPuppetPlanParam returns a new PuppetPlanParam object
func NewPuppetPlanParam() (p *PuppetPlanParam) {
	p = new(PuppetPlanParam)
	// Set Defaults
	return
}

// Init : Create new PatchRun object
func (p *PuppetPlanParam) Init() error {
	err := p.buildTemplate()
	if err != nil {
		return err
	}
	return GetDB().Create(p).Error
}

// Save : Save PatchRun object
func (p *PuppetPlanParam) Save() error {
	err := p.buildTemplate()
	if err != nil {
		return err
	}
	return GetDB().Save(p).Error
}

// Delete : Delete PatchRun object
func (p *PuppetPlanParam) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// No Child Objects yet
	// 	// TODO: DO something about the template?
	// }
	return GetDB().Delete(p).Error
}

// GetTemplate : Returns Template for Param
func (p *PuppetPlanParam) GetTemplate() (tpl *template.Template, err error) {
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
func (p *PuppetPlanParam) GetDefaultValue() (value string, err error) {
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
func (p *PuppetPlanParam) buildTemplate() (err error) {
	templateName := fmt.Sprintf("PuppetPlanParamID-%v", p.ID) // I am not even sure where this us used.
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

// GetPuppetPlanParamByID returns patch run object by ID
func GetPuppetPlanParamByID(id uint) (p *PuppetPlanParam, err error) {
	p = new(PuppetPlanParam)
	err = GetDB().First(p, id).Error
	return
}

// GetPuppetPlanParams returns a list of all PuppetPlanParams
func GetPuppetPlanParams() (params []*PuppetPlanParam) {
	params = make([]*PuppetPlanParam, 0)
	GetDB().Order("name").Find(&params)
	return
}
