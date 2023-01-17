package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func main() {

	data := map[string]interface{}{}
	json.Unmarshal([]byte(source), &data)
	t := template.New("test")
	t = t.Funcs(template.FuncMap{
		"add": func(x, y interface{}) int {
			var a, b int
			switch v := x.(type) {
			case int:
				a = v
			case float64:
				a = int(v)
			}
			switch v := y.(type) {
			case int:
				b = v
			case float64:
				b = int(v)
			}
			return a + b
		},
		"div":       func(x, y int) float64 { return float64(x) / float64(y) },
		"hasPrefix": strings.HasPrefix,
		"hasBit":    func(x float64, y int) bool { return (int(x)>>y)&1 == 1 },
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
	fmt.Fprintf(os.Stdout, clean_template)
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
{{ $nAsync := 0 }}
{{ $nJWT := 0 }}
{{ $hasCB := false }}
{{ $hasRL := false }}
{{ $hasLogging := false }}
{{ $nAPIKeys := 0 }}
{{ $hasHTTPSec := false }}
{{ $hasTele := false }}
{{ $hasRouter := false }}
{{ $hasCORS := false }}
{{ $hasBotDetect := false }}
{{ $nTele := 0 }}

{{/*
	Service Bits definition (.d array with one element)
*/}}
{{ $serviceBits := (index .d 0) }}
{{ $hasPlugins := hasBit $serviceBits 0 }}
{{ $hasSequentialStart := hasBit $serviceBits 1 }}
{{ $hasDebug := hasBit $serviceBits 2 }}
{{ $hasInsecureConnections := hasBit $serviceBits 3 }}
{{ $hasDisableREST := hasBit $serviceBits 4 }}
{{ $hasTLSBlock := hasBit $serviceBits 5 }}
{{ $hasTLSEnabled := hasBit $serviceBits 6 }}
{{ $hasMTLS := hasBit $serviceBits 7 }}
{{ $hasSystemCADisabled := hasBit $serviceBits 8 }}
{{ $hasTLSCaCerts := hasBit $serviceBits 9 }}


{{ range $ks,$vs := .c }} {{/* Service configurations */}}
	{{ $hasRL = or $hasRL (eq $ks "qos/ratelimit/router") }}
	{{ $hasLogging = or $hasLogging (eq $ks "telemetry/logging") }}
	{{ $hasHTTPSec = or $hasHTTPSec (eq $ks "security/http") }}
	{{ $hasBotDetect = or $hasBotDetect (eq $ks "security/bot-detector") }}
	{{ $hasRouter = or $hasRouter (eq $ks "router") }}
	{{ $hasCORS = or $hasCORS (eq $ks "security/cors") }}
	{{ $hasTele = or $hasTele (hasPrefix $ks "telemetry/") }}
	{{ if hasPrefix $ks "telemetry/"}}{{$nTele = add $nTele 1}}{{end}}
{{ end }}
{{ if .a }}{{ $nAsync = len .a }}{{end}}
{{ if .e }}
{{ $nEndpoints := len .e }}
{{ $nQueryStrings := 0.0 }}
{{ $nHeadersToPass := 0.0 }}


	{{ range .e }}

	{{ $output_encoding := (index .d 0)}}
	{{ $nQueryStrings = add $nQueryStrings (index .d 1) }}
	{{ $nHeadersToPass = add $nHeadersToPass (index .d 2) }}

		{{ if .b}}
			{{ $nBackends = add $nBackends ( len .b ) }}
		{{else}}
			{{ $ERROR }} There is an endpoint defined without any backends
		{{end}}
		{{ range $ke,$ve := .c }}
			{{ if eq $ke "auth/validator"}}{{$nJWT = add $nJWT 1}}{{end}}
			{{ if eq $ke "auth/api-keys"}}{{$nAPIKeys = add $nAPIKeys 1}}{{end}}
			{{ $hasRL = or $hasRL (eq $ke "qos/ratelimit/router") }}
			{{ $hasRL = or $hasRL (eq $ke "qos/ratelimit/proxy") }}
		{{ end }}
		{{ range .b }}
			{{ range $kb,$vb := .c }}
				{{ $hasCB = or $hasCB (eq $kb "qos/circuit-breaker") }}
				{{ $hasRL = or $hasRL (eq $kb "qos/ratelimit/proxy") }}
			{{ end }}
		{{ end }}
	{{ end }}


	{{/* START TEMPLATE */}}

{{/* SERVICE SETTINGS */}}
========== Service settings ==========
{{ if not $hasCORS }}
	{{ $DEBUG }} You don't have {{$ColorBlue}}security/cors{{$ColorReset}} and clients won't do cross-origin requests.
{{ else }}
	{{ $DEBUG }} CORS is enabled.
{{ end }}
{{ if $hasBotDetect }}
	{{ $DEBUG }} Bot detector is enabled
{{ end }}

{{ if not $hasTele }}
	{{ $WARNING }} Hope you are good reading logs, because you don't have any telemetry system enabled.
{{ else }}
	{{ $DEBUG }} You have {{$nTele}} telemetry component(s) enabled.
{{ end }}
{{ if not $hasLogging }}
	{{ $WARNING }} You don't have the {{$ColorBlue}}telemetry/logging{{$ColorReset}} component enabled, which is essential in any production installation.
{{ end }}
{{ if $hasPlugins }}
{{end}}
{{ if $hasSequentialStart }}
{{end}}
{{ if $hasDebug }}
{{end}}
{{ if $hasInsecureConnections }}
{{end}}
{{ if $hasDisableREST }}
{{end}}
{{ if not $hasTLSBlock }}
	{{$INFO}} You are not using {{$ColorBlue}}tls{{$ColorReset}}. Hopefully you are terminating SSL before KrakenD.
{{ else }}
	{{ if not $hasTLSEnabled }}
		{{$WARNING}} You have configured {{$ColorBlue}}tls{{$ColorReset}} but it's disabled!
	{{end}}
	{{ if $hasMTLS }}
	{{ $DEBUG }} MTLS is configured.
	{{end}}
	{{ if $hasSystemCADisabled }}
	{{ $DEBUG }} The system CA is disabled
	{{end}}
	{{ if $hasTLSCaCerts }}
	{{ $DEBUG }} There are custom CAs for TLS
	{{end}}
{{ end }}
{{ if not $hasHTTPSec}}
	{{$WARNING}} You don't have any {{$ColorBlue}}security/http{{$ColorReset}} option enabled.
{{end}}
{{ if $hasRouter}}
	{{$WARNING}} You have {{$ColorBlue}}router{{$ColorReset}} customizations overriding standard behavior.
{{end}}
{{ if not $hasRL }}
	{{ $WARNING }} You are exposing an All-You-Can-Eat API without any type of stateless rate limiting.
{{ end }}

========== Endpoint configuration ==========
{{ $DEBUG }} There are {{ $nEndpoints }} endpoint(s) configured
{{ $DEBUG }} There are {{ printf "%.2f" (div $nQueryStrings $nEndpoints)}} query strings per endpoint
{{ $DEBUG }} There are {{ printf "%.2f" (div $nHeadersToPass $nEndpoints)}} passing headers per endpoint
{{ if $nBackends }}
		{{ $DEBUG }} There are {{ $nBackends }} backend(s) configured
	{{ end }}

	{{ if and $nBackends $nEndpoints }}
		{{ $avg := div $nBackends $nEndpoints }}
		{{ if lt $avg 1.1}}
		{{$WARNING}} You are not taking advantage of aggregation. There are only {{ printf "%.2f" $avg }} backends per endpoint.
		{{else}}
		{{$DEBUG}} There are {{ printf "%.2f" $avg }} backends per endpoint
		{{end}}
	{{ end }}

	{{ if $nJWT }}
		{{$DEBUG}} You have {{ $nJWT }} endpoint(s) configured with JWT validation.
	{{ else }}
		{{$WARNING}} No endpoint is protected by JWT
	{{ end }}

	{{ if $nAPIKeys }}
		{{$DEBUG}} You have {{ $nAPIKeys }} endpoint(s) requiring API Keys.
	{{ end }}


	{{ if not $hasCB }}
		{{ $WARNING }} Your backends are not protected with {{$ColorBlue}}qos/circuit-breaker{{$ColorReset}}.
	{{ end }}
{{ else }}
	{{ $ERROR }} No endpoints defined!
{{ end}}
========== Async agents ==========
{{ $DEBUG }} There are {{ $nAsync }} Async Agents configured
`

const source = `{

    "d": [
		160
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
