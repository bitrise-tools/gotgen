# gotgen

Minimalistic Go Template based Generator

[![Build Status](https://www.bitrise.io/app/bf4a1f1b34d5648f/status.svg?token=fSHjhJa7ZSUH_61azXd_xg&branch=master)](https://www.bitrise.io/app/bf4a1f1b34d5648f)

## How to use

Run:

```
gotgen init
```

to create a `gg.conf.json` GotGen configuration file in the current directory.

Then run:

```
gotgen generate
```

Reads `gg.conf.json` from the currend directory, parses all the `.gg` files in the current directory,
then runs [Go template](https://golang.org/pkg/text/template/) on all `.gg` file content,
with Inventory defined in the `gg.conf.json` exposed as Inventory for the Go Template,
then saves the generated files with the same name without `.gg` extension.


## Example config and template file

Example `gg.conf.json` config file:

```
{
  "inventory": {
    "KeyBool": true,
    "KeyOne": "value for key one",
    "KeyTwo": 2
  }
}
```

Example `*.gg` template file:

```
This is an example GotGen template file.

If you run "gotgen generate" that will create a new file in this directory, with the filename "example.txt".

The gg.conf.json file in this directory includes the Inventory, which can be used in these .gg Go Templates (https://golang.org/pkg/text/template/).

For example, if you run "gogen generate" now in this directory, the generated "example.txt" file will be generated by including the value of:

- KeyOne here: {{ .KeyOne }}
- and KeyTwo here: {{ .KeyTwo }}

And the following section will be showns based on KeyBool's value:

{{ if .KeyBool }}KeyBool was true{{ else }}KeyBool was false{{ end }}

That's all you need to know.

GG ;)
```
