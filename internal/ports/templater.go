package ports

type Templater interface {
	Render(template string, templateName string, values map[string]interface{}) (string, error)
}
