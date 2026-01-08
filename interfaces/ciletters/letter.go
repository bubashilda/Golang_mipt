//go:build !solution

package ciletters

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed pattern.txt
var pattern string

func MakeLetter(n *Notification) (string, error) {
	var ans strings.Builder

	funcMap := template.FuncMap{
		"suffix": func(job Job) string {
			output := strings.Split(job.RunnerLog, "\n")
			return strings.TrimSpace(strings.Join(output[max(0, len(output)-10):], "\n"+strings.Repeat(" ", 12)))
		},
	}

	tmpl, err := template.New("letter").Funcs(funcMap).Parse(pattern)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(&ans, n)
	return ans.String(), err
}
