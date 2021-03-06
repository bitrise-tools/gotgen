package cmd

import (
	"encoding/json"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/gotgen/configs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// write base config into file
		conf := configs.Model{
			Inventory: map[string]interface{}{
				"KeyOne":  "value for key one",
				"KeyTwo":  2,
				"KeyBool": true,
				"Nested": map[string]interface{}{
					"KeyA": map[string]interface{}{
						"Key1": "KeyA-Key1 value",
					},
				},
			},
			Delimiter: configs.DelimiterModel{
				Left:  "{{",
				Right: "}}",
			},
		}

		jsonCont, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			return errors.Wrap(err, "Failed to generate config JSON")
		}

		confFilePth := "./" + gotgenConfigFileName
		if err := fileutil.WriteBytesToFile(confFilePth, jsonCont); err != nil {
			return errors.Wrapf(err, "Failed to write config into file (to path: %s)", confFilePth)
		}

		// write a test .gg file
		exampleFilePth := "./example.txt.gg"
		if err := fileutil.WriteStringToFile(
			exampleFilePth,
			`This is an example GotGen template file.

If you run "gotgen generate" that will create a new file in this directory, with the filename "example.txt".

The gg.conf.json file in this directory includes the Inventory, which can be used in these .gg Go Templates (https://golang.org/pkg/text/template/).

For example, if you run "gotgen generate" now in this directory, the generated "example.txt" file will be generated by including the value of:

- KeyOne here: {{ var "KeyOne" }}
- and KeyTwo here: {{ .KeyTwo }}
- and Nested.KeyA.Key1 here: {{ .Nested.KeyA.Key1 }}

And the following section will be showns based on KeyBool's value:

{{ if .KeyBool }}KeyBool was true{{ else }}KeyBool was false{{ end }}

Environment variables can also be included, with {{ getenv "ENV_VAR_KEY" }},
which will result in an empty string if ENV_VAR_KEY is not set. If you want the template
to fail if the env var isn't set use getenvRequired instead.

Inline yaml embedding:
{{ .Nested | yaml }}

Indentation:
{{ "a\nb\n" | indentWithSpaces 4 }}

Some math:
add: {{ 6 | add 2 }}
subtract: {{ 6 | subtract 2 }}
multiply: {{ 6 | multiply 2 }}
divide: {{ 6 | divide 2 }}
modulo: {{ 6 | modulo 2 }}

That's all you need to know.

GG ;)
`,
		); err != nil {
			return errors.Wrapf(err, "Failed to write example .gg file (to path: %s)", exampleFilePth)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
