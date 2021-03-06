package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"encoding/json"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/templateutil"
	"github.com/bitrise-io/gotgen/configs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	ggTemplateFilePathFlag = ""
	outputFilePathFlag     = ""
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	RunE: generate,
}

func init() {
	RootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	generateCmd.Flags().StringVar(&ggTemplateFilePathFlag, "file", "", ".gg file path - if specified only this single .gg file will be used as input, instead of scanning the whole directory for .gg files")
	generateCmd.Flags().StringVar(&outputFilePathFlag, "out-file-path", "", "Output file path")
}

func generate(cmd *cobra.Command, args []string) error {
	// Read Inventory
	log.Println(colorstring.Blue("Reading GotGen config ..."))
	ggConfContent, err := fileutil.ReadBytesFromFile(gotgenConfigFileName)
	if err != nil {
		return errors.Wrapf(err, "Failed to read GotGen config (%s) file", gotgenConfigFileName)
	}
	ggConf := configs.Model{}
	if err := json.Unmarshal(ggConfContent, &ggConf); err != nil {
		return errors.Wrap(err, "Failed to parse GotGen config (JSON)")
	}
	log.Println(colorstring.Green("[DONE] Reading GotGen config"))

	//
	templateFiles := map[string]string{}
	if len(ggTemplateFilePathFlag) > 0 {
		log.Println(colorstring.Blue("Using only the specified template file: " + ggTemplateFilePathFlag))
		oFilePth := outputFilePathFlag
		if len(oFilePth) < 1 {
			if !strings.HasSuffix(ggTemplateFilePathFlag, ".gg") {
				return errors.Errorf("If you specify an input file path that either has to be a .gg file (.gg extension) or also specify the out-file-path option to specify where the generated output should be stored")
			}
			oFilePth = strings.TrimSuffix(ggTemplateFilePathFlag, ".gg")
		}
		templateFiles[ggTemplateFilePathFlag] = oFilePth
	} else {
		log.Println(colorstring.Blue("Searching for templates ..."))
		files, err := filepath.Glob("*.gg")
		if err != nil {
			return errors.Wrap(err, "Failed to scan .gg template files")
		}
		for _, aFilePath := range files {
			templateFiles[aFilePath] = strings.TrimSuffix(aFilePath, ".gg")
		}
	}

	if len(templateFiles) < 1 {
		return errors.Errorf("No template file specified or found.")
	}

	log.Println(colorstring.Blue("Generating ..."))
	fmt.Println()
	for aTemplatePth, aOutputFilePath := range templateFiles {
		if err := generateFileForTemplate(aTemplatePth, aOutputFilePath, ggConf); err != nil {
			return errors.WithStack(err)
		}
	}
	fmt.Println()
	log.Println(colorstring.Green("[DONE] Searching for templates and generating files"))
	fmt.Println()

	return nil
}

func generateFileForTemplate(templatePath, generatedFilePath string, ggconf configs.Model) error {
	fmt.Println(" * ", templatePath, " => ", generatedFilePath)

	templateCont, err := fileutil.ReadStringFromFile(templatePath)
	if err != nil {
		return errors.Wrapf(err, "Failed to read template content (path: %s)", templatePath)
	}

	generatedContent, err := generateContent(templateCont, ggconf.Inventory, ggconf.Delimiter.Left, ggconf.Delimiter.Right)
	if err != nil {
		return errors.Wrapf(err, "Failed to generate file based on content (%s) - invalid content?", templatePath)
	}

	if err := fileutil.WriteStringToFile(generatedFilePath, generatedContent); err != nil {
		return errors.Wrapf(err, "Failed to write generated content into file (to path: %s)", generatedFilePath)
	}
	fmt.Println("   ", colorstring.Green("[OK]"))

	return nil
}

func generateContent(templateCont string, inventory map[string]interface{}, delimiterLeft, delimiterRight string) (string, error) {
	generatedContent, err := templateutil.EvaluateTemplateStringToStringWithDelimiterAndOpts(
		templateCont,
		inventory, createAvailableTemplateFunctions(inventory),
		delimiterLeft, delimiterRight,
		[]string{"missingkey=error"})
	if err != nil {
		return "", errors.WithStack(err)
	}
	return generatedContent, nil
}

func createAvailableTemplateFunctions(inventory map[string]interface{}) template.FuncMap {
	return template.FuncMap{
		"var": func(key string) (interface{}, error) {
			val, isFound := inventory[key]
			if !isFound {
				return "", errors.Errorf("No value found for key: %s", key)
			}
			return val, nil
		},
		"getenv": func(key string) string {
			return os.Getenv(key)
		},
		"getenvRequired": func(key string) (string, error) {
			if val := os.Getenv(key); len(val) > 0 {
				return val, nil
			}
			return "", errors.Errorf("No environment variable value found for key: %s", key)
		},
		"yaml":             yamlFn,
		"indentWithSpaces": indentWithSpaces,
		"add":              add,
		"subtract":         subtract,
		"multiply":         multiply,
		"divide":           divide,
		"modulo":           modulo,
	}
}

// ------------------------------------------------------------
// Utilify functions
// ------------------------------------------------------------

func yamlFn(obj interface{}) (string, error) {
	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return "", errors.Errorf("Failed to generate yaml for object, error: %s", err)
	}
	return string(bytes), nil
}

func indentWithSpaces(indentSpaceCharCount int, s string) string {
	if len(s) < 1 {
		return ""
	}

	indentationString := strings.Repeat(" ", indentSpaceCharCount)

	lines := strings.SplitAfter(s, "\n")
	return indentationString + strings.Join(lines, indentationString)
}

// ------------------------------------------------------------
// Arithmetic functions
// Based on https://github.com/hashicorp/consul-template/blob/9ef7c22f1ec0540ef746d0ecf873353ae57c4e77/template/funcs.go#L956
// ------------------------------------------------------------

// add returns the sum of a and b.
// Based on https://github.com/hashicorp/consul-template/blob/9ef7c22f1ec0540ef746d0ecf873353ae57c4e77/template/funcs.go#L956
func add(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() + int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) + bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() + bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() + float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() + float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("add: unknown type for %q (%T)", av, a)
	}
}

// subtract returns the difference of b from a.
// Based on https://github.com/hashicorp/consul-template/blob/9ef7c22f1ec0540ef746d0ecf873353ae57c4e77/template/funcs.go#L956
func subtract(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() - bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() - int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) - bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() - bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() - float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() - float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("subtract: unknown type for %q (%T)", av, a)
	}
}

// multiply returns the product of a and b.
// Based on https://github.com/hashicorp/consul-template/blob/9ef7c22f1ec0540ef746d0ecf873353ae57c4e77/template/funcs.go#L956
func multiply(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() * bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() * int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) * bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() * bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() * float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() * float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("multiply: unknown type for %q (%T)", av, a)
	}
}

// divide returns the division of b from a.
// Based on https://github.com/hashicorp/consul-template/blob/9ef7c22f1ec0540ef746d0ecf873353ae57c4e77/template/funcs.go#L956
func divide(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() / bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() / int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) / bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() / bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() / float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() / float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("divide: unknown type for %q (%T)", av, a)
	}
}

// modulo returns the modulo of b from a.
// Based on https://github.com/hashicorp/consul-template/blob/9ef7c22f1ec0540ef746d0ecf873353ae57c4e77/template/funcs.go#L956
func modulo(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() % bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() % int64(bv.Uint()), nil
		default:
			return nil, fmt.Errorf("modulo: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) % bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() % bv.Uint(), nil
		default:
			return nil, fmt.Errorf("modulo: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("modulo: unknown type for %q (%T)", av, a)
	}
}
