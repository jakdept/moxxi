package moxxiConf

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFormHandler(t *testing.T) {
	var testData = []struct {
		reqMethod string
		reqURL    string
		reqBody   string
		resBody   string
		resCode   int
		fileOut   string
	}{
		{
			reqMethod: "POST",
			reqURL:    "domain.com",
			reqBody:   ``,
			resBody:   ``,
			resCode:   200,
			fileOut:   ``,
		},
	}

	templPath := os.TempDir()
	templExt := ".out"
	baseURL := "domain.com"
	subdomainLen := 1
	excludes := []string{"a.domain.com", "b.domain.com", "c.domain.com"}
	templString := `{{.IntHost}} {{.IntIP}} {{.IntPort}} {{.Encrypted}} {{ range .StripHeaders }}{{.}} {{end}}`
	templ := template.Must(template.New("testing").Parse(templateString))

}
