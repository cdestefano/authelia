package templates

import (
	"fmt"
	th "html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
	tt "text/template"
)

const (
	envPrefix  = "AUTHELIA_"
	envXPrefix = "X_AUTHELIA_"
)

// IMPORTANT: This is a copy of github.com/authelia/authelia/internal/configuration's secretSuffixes except all uppercase.
// Make sure you update these at the same time.
var envSecretSuffixes = []string{
	"KEY", "SECRET", "PASSWORD", "TOKEN", "CERTIFICATE_CHAIN",
}

func isSecretEnvKey(key string) (isSecretEnvKey bool) {
	key = strings.ToUpper(key)

	if !strings.HasPrefix(key, envPrefix) && !strings.HasPrefix(key, envXPrefix) {
		return false
	}

	for _, s := range envSecretSuffixes {
		suffix := strings.ToUpper(s)

		if strings.HasSuffix(key, suffix) {
			return true
		}
	}

	return false
}

func templateExists(path string) (exists bool) {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

func readTemplate(name, ext, category, overridePath string) (tPath string, embed bool, data []byte, err error) {
	if overridePath != "" {
		tPath = filepath.Join(overridePath, name+ext)

		if templateExists(tPath) {
			if data, err = os.ReadFile(tPath); err != nil {
				return tPath, false, nil, fmt.Errorf("failed to read template override at path '%s': %w", tPath, err)
			}

			return tPath, false, data, nil
		}
	}

	tPath = path.Join("src", category, name+ext)

	if data, err = embedFS.ReadFile(tPath); err != nil {
		return tPath, true, nil, fmt.Errorf("failed to read embedded template '%s': %w", tPath, err)
	}

	return tPath, true, data, nil
}

func parseTextTemplate(name, tPath string, embed bool, data []byte) (t *tt.Template, err error) {
	if t, err = tt.New(name + extText).Funcs(FuncMap()).Parse(string(data)); err != nil {
		if embed {
			return nil, fmt.Errorf("failed to parse embedded template '%s': %w", tPath, err)
		}

		return nil, fmt.Errorf("failed to parse template override at path '%s': %w", tPath, err)
	}

	return t, nil
}

func parseHTMLTemplate(name, tPath string, embed bool, data []byte) (t *th.Template, err error) {
	if t, err = th.New(name + extHTML).Funcs(FuncMap()).Parse(string(data)); err != nil {
		if embed {
			return nil, fmt.Errorf("failed to parse embedded template '%s': %w", tPath, err)
		}

		return nil, fmt.Errorf("failed to parse template override at path '%s': %w", tPath, err)
	}

	return t, nil
}

func loadEmailTemplate(name, overridePath string) (t *EmailTemplate, err error) {
	var (
		embed bool
		tpath string
		data  []byte
	)

	t = &EmailTemplate{}

	if tpath, embed, data, err = readTemplate(name, extText, TemplateCategoryNotifications, overridePath); err != nil {
		return nil, err
	}

	if t.Text, err = parseTextTemplate(name, tpath, embed, data); err != nil {
		return nil, err
	}

	if tpath, embed, data, err = readTemplate(name, extHTML, TemplateCategoryNotifications, overridePath); err != nil {
		return nil, err
	}

	if t.HTML, err = parseHTMLTemplate(name, tpath, embed, data); err != nil {
		return nil, err
	}

	return t, nil
}
