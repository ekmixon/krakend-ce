package main

import (
	"encoding/json"
	"strings"
	"text/template"
	"fmt"
	"bytes"
	"regexp"
	"os"
)

func main() {

	data := map[string]interface{}{}
	json.Unmarshal([]byte(source), &data)
	t := template.New("test")
	t = t.Funcs(template.FuncMap{
		"add":       func(x, y int) int { return x + y },
		"div":       func(x, y int) float64 { return float64(x) / float64(y) },
		"hasPrefix": strings.HasPrefix,
		"hasBit":    func(x float64, y int) bool { return (int(x) >> y) == 1 },
	})
	var err error
	t, err = t.Parse(tmpl)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	// Remove tabs and extra new lines. Doing it in the template it's kind of hard to mantain
	clean_template := regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(buf.String()), "\n")
	fmt.Fprintf(os.Stdout,clean_template)
}

// Define a template.
const tmpl = `
{{ $ColorGreen := "\033[32;1m" }}
{{ $ColorWhite := "\033[37;1m" }}
{{ $ColorYellow := "\033[33;1m" }}
{{ $ColorRed := "\033[31;1m" }}
{{ $ColorBlue := "\033[34;1m" }}
{{ $ColorMagenta := "\033[36;1m" }}
{{ $ColorCyan := "\033[36;1m" }}
{{ $ColorReset := "\033[0m" }}
{{ $CRITICAL := printf "%s[CRITICAL]%s " $ColorMagenta $ColorReset }}
{{ $ERROR := printf "%s[ERROR]%s " $ColorRed $ColorReset }}
{{ $WARNING := printf "%s[WARNING]%s " $ColorYellow $ColorReset }}
{{ $INFO := printf "%s[INFO]%s " $ColorGreen $ColorReset }}
{{ $DEBUG := printf "%s[DEBUG]%s " $ColorCyan $ColorReset }}
{{ $nBackends := 0 }}
{{ $hasJWT := false }}
{{ $hasCB := false }}
{{ $hasRL := false }}
{{ $hasLogging := false }}
{{ $hasAPIKeys := false }}
{{ $hasTele := false }}
{{ $nTele := 0 }}
{{ $bitTLSEnabled := 5 }}
{{ range $ks,$vs := .c }}
	{{ $hasRL = or $hasRL (eq $ks "qos/ratelimit/router") }}
	{{ $hasLogging = or $hasLogging (eq $ks "telemetry/logging") }}
	{{ $hasTele = or $hasTele (hasPrefix $ks "telemetry/") }}
	{{ if hasPrefix $ks "telemetry/"}}{{$nTele = add $nTele 1}}{{end}}
{{ end }}
{{ if .e }}
{{ $nEndpoints := len .e }}
	{{ range .e }}
		{{ $nBackends = add $nBackends ( len .b ) }}
		{{ range $ke,$ve := .c }}
			{{ $hasJWT = or $hasJWT (eq $ke "auth/validator") }}
			{{ $hasRL = or $hasRL (eq $ke "qos/ratelimit/router") }}
			{{ $hasRL = or $hasRL (eq $ke "qos/ratelimit/proxy") }}
			{{ $hasAPIKeys = or $hasAPIKeys (eq $ke "auth/api-keys") }}
		{{ end }}
		{{ range .b }}
			{{ range $kb,$vb := .c }}
				{{ $hasCB = or $hasCB (eq $kb "qos/circuit-breaker") }}
				{{ $hasRL = or $hasRL (eq $kb "qos/ratelimit/proxy") }}
			{{ end }}
		{{ end }}
	{{ end }}


	{{/* START TEMPLATE */}}

	{{ $DEBUG }} There are {{ $nEndpoints }} endpoints configured

	{{ if $nBackends }}
		{{ $DEBUG }} There are {{ $nBackends }} backends configured
	{{ end }}

	{{ if and $nBackends $nEndpoints }}
		{{ $avg := div $nBackends $nEndpoints }}
		{{ if lt $avg 1.3}}
		{{$WARNING}} You are not taking advantage of aggregation. There are only {{ printf "%.2f" $avg }} backends per endpoint.
		{{else}}
		{{$DEBUG}} There are {{ printf "%.2f" $avg }} backends per endpoint
		{{end}}
	{{ end }}

	{{ if $hasJWT }}
		{{$DEBUG}} You have endpoints configured with JWT validation.
	{{ else }}
		{{$WARNING}} No endpoint is protected by JWT
	{{ end }}

	{{ if $hasAPIKeys }}
		{{$DEBUG}} You have endpoints requiring API Keys.
	{{ end }}


	{{ if not $hasCB }}
		{{ $WARNING}} Your backends are not protected with {{$ColorBlue}}qos/circuit-breaker{{$ColorReset}}.
	{{ end }}

	{{ if not $hasRL }}
		{{ $WARNING}} You are exposing an All-You-Can-Eat API without any type of stateless rate limiting.
	{{ end }}

{{ else }}
	{{ $ERROR }} No endpoints defined!
{{ end}}
{{/* SERVICE SETTINGS */}}
{{ if not $hasTele }}
	{{ $WARNING}} Hope you are good reading logs, because you don't have any telemetry system enabled.
{{ else }}
	{{ $DEBUG}} You have {{$nTele}} telemetry component(s) enabled.
{{ end }}
{{ if not $hasLogging }}
	{{ $WARNING}} You don't have the {{$ColorBlue}}telemetry/logging{{$ColorReset}} component enabled, which is essential in any production installation.
{{ end }}
{{ if not (ge (index .d 0) 32.0) }}
	{{$INFO}} You are not using TLS. Hopefully you are terminating SSL before KrakenD.
{{ else if not (hasBit (index .d 0) $bitTLSEnabled) }}
	{{$WARNING}} You have configured TLS but it's disabled.
{{ end }}
`

const source = `{

    "d": [
		64
	],
	"a": [],
	"e": [
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {
				"auth/api-keys": []
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						96
					],
					"c": {}
				},
				{
					"d": [
						320
					],
					"c": {}
				}
			],
			"c": {
				"proxy": [
					1
				]
			}
		},
		{
			"d": [
				1,
				0,
				0
			],
			"b": [
				{
					"d": [
						1
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				1,
				0,
				0
			],
			"b": [
				{
					"d": [
						1
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				32,
				0,
				0
			],
			"b": [
				{
					"d": [
						32
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				32,
				0,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				64,
				0,
				0
			],
			"b": [
				{
					"d": [
						32
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						128
					],
					"c": {}
				}
			],
			"c": {
				"proxy": [
					16
				]
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				},
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {
				"proxy": [
					16
				]
			}
		},
		{
			"d": [
				2,
				0,
				1
			],
			"b": [
				{
					"d": [
						320
					],
					"c": {
						"validation/cel": []
					}
				},
				{
					"d": [
						320
					],
					"c": {
						"validation/cel": []
					}
				}
			],
			"c": {
				"validation/cel": []
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {
				"auth/validator": [],
				"validation/cel": []
			}
		},
		{
			"d": [
				2,
				0,
				2
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				2
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				1
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				1,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				1,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						96
					],
					"c": {}
				},
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {
				"proxy": [
					1
				]
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						96
					],
					"c": {}
				},
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {
				"proxy": [
					1
				]
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						1088
					],
					"c": {
						"proxy": [
							2
						]
					}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						1216
					],
					"c": {}
				},
				{
					"d": [
						96
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						304
					],
					"c": {
						"qos/circuit-breaker": [],
						"qos/http-cache": []
					}
				},
				{
					"d": [
						304
					],
					"c": {
						"qos/circuit-breaker": [],
						"qos/http-cache": []
					}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						320
					],
					"c": {}
				},
				{
					"d": [
						320
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {
						"backend/http": []
					}
				},
				{
					"d": [
						64
					],
					"c": {
						"backend/http": []
					}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						128
					],
					"c": {}
				}
			],
			"c": {
				"auth/validator": [],
				"proxy": [
					16
				]
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						128
					],
					"c": {}
				}
			],
			"c": {
				"auth/validator": [],
				"proxy": [
					16
				]
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						128
					],
					"c": {}
				}
			],
			"c": {
				"auth/signer": [],
				"proxy": [
					16
				]
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						1088
					],
					"c": {
						"modifier/lua-backend": []
					}
				}
			],
			"c": {}
		},
		{
			"d": [
				2,
				0,
				1
			],
			"b": [
				{
					"d": [
						1088
					],
					"c": {
						"modifier/lua-backend": []
					}
				}
			],
			"c": {
				"modifier/lua-proxy": []
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						320
					],
					"c": {}
				}
			],
			"c": {
				"modifier/lua-proxy": []
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						320
					],
					"c": {}
				}
			],
			"c": {
				"modifier/lua-proxy": []
			}
		},
		{
			"d": [
				2,
				0,
				0
			],
			"b": [
				{
					"d": [
						64
					],
					"c": {}
				}
			],
			"c": {
				"validation/json-schema": []
			}
		}
	],
	"c": {
		"router": [
			16
		],
		"security/bot-detector": [
			2,
			2,
			2,
			0
		],
		"security/cors": [],
		"telemetry/logging": [],
		"telemetry/metrics": []
	}
}`
