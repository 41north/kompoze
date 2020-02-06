package internal

import (
	"encoding/json"
	"github.com/Masterminds/sprig"
	"os"
	"path/filepath"
	"syscall"
	tpl "text/template"

	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
	"github.com/qri-io/jsonschema"
	log "github.com/sirupsen/logrus"
)

// Render renders a concrete definition file
func Render(definitionPath string, basePath string, delims []string, forceStdOut bool, noOverwrite bool) {
	definition := newDefinition(absPath(basePath, definitionPath))

	resolveIncludeVars(basePath, definition.Vars.Include, definition.Vars.Global)

	for _, template := range definition.Templates {
		resolveIncludeVars(basePath, template.IncludeVars, template.LocalVars)

		if err := mergo.Merge(template.LocalVars, definition.Vars.Global, mergo.WithOverride); err != nil {
			log.Fatalf("Problem merging variables: %s", err)
		}

		log.Infof("Rendering template: %s", template.String())
		renderTpl(template, basePath, delims, forceStdOut, noOverwrite)
	}

	log.Infof("Finished rendering templates!")
}

func validateIncludeVars(v *includeVars) []jsonschema.ValError {
	rs := &jsonschema.RootSchema{}
	if err := json.Unmarshal(includeVarsSchemaData, rs); err != nil {
		log.Fatalf("Error un-marshaling schema: %s", err)
	}

	log.Infof("Validating include vars file: %s", v.String())

	j := v.MarshalJSON()
	if errs, _ := rs.ValidateBytes(j); len(errs) > 0 {
		return errs
	}

	return nil
}

func resolveIncludeVars(basePath string, include *[]string, global *map[string]interface{}) {
	if include != nil && len(*include) > 0 {
		for _, path := range *include {
			path = absPath(basePath, path)

			v := &includeVars{}
			if _, err := toml.DecodeFile(path, v); err != nil {
				log.Fatalf("Problem decoding TOML vars file: %s", err)
			}

			log.Infof("Loaded global vars file: %s", v.String())
			validateIncludeVars(v)

			if err := mergo.Merge(global, v.Vars, mergo.WithOverride); err != nil {
				log.Fatalf("Problem merging variables: %s", err)
			}
		}
	}
}

func absPath(basePath string, path string) string {
	if !filepath.IsAbs(path) {
		return filepath.Join(basePath, path)
	}
	return path
}

func ensureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, err := os.Stat(dirName); err != nil {
		er := os.MkdirAll(dirName, os.ModePerm)
		if er != nil {
			log.Fatalf("unable to create directory: %s. Error: %s", fileName, er)
		}
	}
}

func renderTpl(t template, basePath string, delims []string, noOverwrite bool, forceStdOut bool) bool {
	t.Src = absPath(basePath, t.Src)
	if t.Dest != "" {
		t.Dest = absPath(basePath, t.Dest)
	}

	customFns := tpl.FuncMap{
		"exists":   exists,
		"parseUrl": parseURL,
		"isTrue":   isTrue,
		"isFalse":  isFalse,
		"loop":     loop,
	}
	sprigFns := tpl.FuncMap(sprig.FuncMap())
	tmpl := tpl.New(filepath.Base(t.Dest)).Funcs(customFns).Funcs(sprigFns).Delims(delims[0], delims[1])

	tmpl, err := tmpl.ParseFiles(t.Src)
	if err != nil {
		log.Fatalf("unable to parse template: %s", err)
	}

	if _, err := os.Stat(t.Dest); err == nil && noOverwrite {
		log.Warnf("File already exists! Ignoring overwrite")
		return false
	}

	dest := os.Stdout
	if !forceStdOut && t.Dest != "" {
		ensureDir(t.Dest)
		if dest, err = os.Create(t.Dest); err != nil {
			log.Fatalf("unable to %s", err)
		}
		defer dest.Close()
	}

	if err = tmpl.ExecuteTemplate(dest, filepath.Base(t.Src), &t.LocalVars); err != nil {
		log.Fatalf("template error: %s\n", err)
	}

	if fi, err := os.Stat(t.Dest); err == nil {
		if err := dest.Chmod(fi.Mode()); err != nil {
			log.Fatalf("unable to chmod temp file: %s\n", err)
		}
		if err := dest.Chown(int(fi.Sys().(*syscall.Stat_t).Uid), int(fi.Sys().(*syscall.Stat_t).Gid)); err != nil {
			log.Fatalf("unable to chown temp file: %s\n", err)
		}
	}

	return true
}
