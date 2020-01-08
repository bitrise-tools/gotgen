# gotgen

Minimalistic Go Template based Generator.

[![Build Status](https://app.bitrise.io/app/bf4a1f1b34d5648f/status.svg?token=fSHjhJa7ZSUH_61azXd_xg&branch=master)](https://app.bitrise.io/app/bf4a1f1b34d5648f)

## How to use

Install:

```shell
go get -u github.com/bitrise-io/gotgen
```

Init:

```shell
gotgen init
```

to create a `gg.conf.json` GotGen configuration file in the current directory.

Then run:

```shell
gotgen generate
```

Reads `gg.conf.json` from the currend directory, parses **all** the `.gg` files in the current directory,
then runs [Go template](https://golang.org/pkg/text/template/) on all `.gg` file content,
with Inventory defined in the `gg.conf.json` exposed as Inventory for the Go Template,
then saves the generated files with the same name without `.gg` extension.

In addition to what's available in the standard Go template package `gotgen` adds a few additional utility functions you can use in your `.gg` templates. For the complete list see the `cmd/generate.go` file's `createAvailableTemplateFunctions` function. A few examples:

- `var`: `{{ var "KeyID" }}`: Fail if KeyID isn't specified in the inventory. Otherwise it works the same as `{{ .KeyID }}` would.
- `getenv`: `{{ getenv "ENV_VAR_KEY" }}`: Get the value of `ENV_VAR_KEY` env var. If the env var does not exist it'll result in an empty string, just like Go's `os.Getenv`.
- `getenvRequired`: `{{ getenvRequired "ENV_VAR_KEY" }}`: Same as `getenv` but it will fail if the env var isn't set or if its value is an empty string.
- `yaml`: `{{ obj | yaml }}`: Generates yaml string for the provided object.
- `indentWithSpaces`: `{{ "some\n multiline\n text" | indentWithSpaces 4 }}`: Indents the specified string with the number of spaces you provide.
- `add`, `subtract`, `multiply`, `divide`, `modulo`: `{{ 6 | add 2 }}`: Simple arithmetic functions.

## Example config and template file

Example `gg.conf.json` config file:

```json
{
  "inventory": {
    "KeyBool": true,
    "KeyOne": "value for key one",
    "KeyTwo": 2,
    "Nested": {
      "KeyA": {
        "Key1": "KeyA-Key1 value"
      }
    }
  },
  "delimiter": {
    "left": "{{",
    "right": "}}"
  }
}
```

Example template file (`example.txt.gg`):

```text
This is an example GotGen template file.

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
```
