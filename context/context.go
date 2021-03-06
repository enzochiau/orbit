/*
Package context helps to populate the application context.

The main goal of the application context is to gather all the data which will be applied to a data-driven template.
*/
package context

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/gulien/orbit/helpers"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type (
	// OrbitContext contains the data necessary for executing a data-driven template.
	OrbitContext struct {
		// TemplateFilePath is the path of the data-driven template.
		TemplateFilePath string

		// Values map contains data from YAML files.
		Values map[string]interface{}

		// EnvFiles map contains pairs from .env files.
		EnvFiles map[string]map[string]string

		// Os is the OS name at runtime.
		Os string
	}

	// OrbitFileMap represents a value given to some flags of generate and run commands.
	// Flags: -v --values, -e --env
	// Value format: name,path;name,path;...
	OrbitFileMap struct {
		// Name is the given name of the file.
		Name string

		// Path is the path of the file.
		Path string
	}
)

// NewOrbitContext instantiates a new OrbitContext.
func NewOrbitContext(templateFilePath string, valuesFiles string, envFiles string) (*OrbitContext, error) {
	// as the data-driven template is mandatory, we must check its validity.
	if templateFilePath == "" || !helpers.FileExists(templateFilePath) {
		return nil, fmt.Errorf("template file \"%s\" does not exist", templateFilePath)
	}

	// let's instantiates our OrbitContext!
	ctx := &OrbitContext{
		TemplateFilePath: templateFilePath,
		Os:               runtime.GOOS,
	}

	// checks if files with values have been specified.
	if valuesFiles != "" {
		data, err := getValuesMap(valuesFiles)
		if err != nil {
			return nil, err
		}

		ctx.Values = data
	}

	// checks if .env files have been specified.
	if envFiles != "" {
		data, err := getEnvFilesMap(envFiles)
		if err != nil {
			return nil, err
		}

		ctx.EnvFiles = data
	}

	return ctx, nil
}

// getValuesMap retrieves values from YAML files.
func getValuesMap(valuesFiles string) (map[string]interface{}, error) {
	filesMap, err := getFilesMap(valuesFiles)
	if err != nil {
		return nil, err
	}

	valuesMap := make(map[string]interface{})
	for _, f := range filesMap {
		// first, checks if the file exists
		if !helpers.FileExists(f.Path) {
			return nil, fmt.Errorf("values file \"%s\" does not exist", f.Path)
		}

		// alright, let's read it to retrieve its data!
		data, err := ioutil.ReadFile(f.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to read the values file \"%s\":\n%s", f.Path, err)
		}

		// last but not least, parses the YAML.
		var values interface{}
		if err := yaml.Unmarshal(data, &values); err != nil {
			return nil, fmt.Errorf("unable to parse the values file \"%s\":\n%s", f.Path, err)
		}

		valuesMap[f.Name] = values
	}

	return valuesMap, nil
}

// getEnvFilesMap retrieves pairs from .env files.
func getEnvFilesMap(envFiles string) (map[string]map[string]string, error) {
	filesMap, err := getFilesMap(envFiles)
	if err != nil {
		return nil, err
	}

	envFilesMap := make(map[string]map[string]string)
	for _, f := range filesMap {
		// first, checks if the file exists
		if !helpers.FileExists(f.Path) {
			return nil, fmt.Errorf("env file \"%s\" does not exist", f.Path)
		}

		// then parses the .env file to retrieve pairs.
		envFilesMap[f.Name], err = godotenv.Read(f.Path)
		if err != nil {
			return nil, fmt.Errorf("unable to parse the env file \"%s\":\n%s", f.Path, err)
		}
	}

	return envFilesMap, nil
}

// getFilesMap reads a string and populates an array of OrbitFileMap instances.
func getFilesMap(s string) ([]*OrbitFileMap, error) {
	var filesMap []*OrbitFileMap

	// checks if the given string is a map of files:
	// if not, considers the string as a path.
	// otherwise tries to populate an array of OrbitFileMap instances.
	parts := strings.Split(s, ";")
	if len(parts) == 1 && len(strings.Split(s, ",")) == 1 {
		filesMap = append(filesMap, &OrbitFileMap{
			Name: "default",
			Path: s,
		})
	} else {
		for _, part := range parts {
			data := strings.Split(part, ",")
			if len(data) != 2 {
				return filesMap, fmt.Errorf("unable to process the files map \"%s\"", s)
			}

			filesMap = append(filesMap, &OrbitFileMap{
				Name: data[0],
				Path: data[1],
			})
		}
	}

	return filesMap, nil
}
